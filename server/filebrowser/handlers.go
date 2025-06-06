package filebrowser

import (
	"archive/zip"
	"encoding/base64"
	"encoding/json"
	"io"
	"io/fs"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"sort"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/marcopiovanello/yt-dlp-web-ui/v3/server/config"
	"github.com/marcopiovanello/yt-dlp-web-ui/v3/server/internal"
	"github.com/marcopiovanello/yt-dlp-web-ui/v3/server/rest" // Import for RespondWithErrorJSON
)

/*
	File based operation handlers (should be moved to rest/handlers.go) or in
	a entirely self-contained package
*/

var (
	videoRe = regexp.MustCompile(`(?i)/\.mov|\.mp4|\.webm|\.mvk|/gmi`)
)

func isVideo(d fs.DirEntry) bool {
	return videoRe.MatchString(d.Name())
}

func isValidEntry(d fs.DirEntry) bool {
	return !strings.HasPrefix(d.Name(), ".") &&
		!strings.HasSuffix(d.Name(), ".part") &&
		!strings.HasSuffix(d.Name(), ".ytdl")
}

type DirectoryEntry struct {
	Name        string    `json:"name"`
	Path        string    `json:"path"`
	Size        int64     `json:"size"`
	ModTime     time.Time `json:"modTime"`
	IsVideo     bool      `json:"isVideo"`
	IsDirectory bool      `json:"isDirectory"`
}

func walkDir(root string) (*[]DirectoryEntry, error) {
	dirs, err := os.ReadDir(root)
	if err != nil {
		return nil, err
	}

	var files []DirectoryEntry

	for _, d := range dirs {
		if !isValidEntry(d) {
			continue
		}

		path := filepath.Join(root, d.Name())

		info, err := d.Info()
		if err != nil {
			return nil, err
		}

		files = append(files, DirectoryEntry{
			Path:        path,
			Name:        d.Name(),
			Size:        info.Size(),
			IsVideo:     isVideo(d),
			IsDirectory: d.IsDir(),
			ModTime:     info.ModTime(),
		})
	}

	return &files, err
}

type ListRequest struct {
	SubDir  string `json:"subdir"`
	OrderBy string `json:"orderBy"`
}

func ListDownloaded(w http.ResponseWriter, r *http.Request) {
	root := config.Instance().DownloadPath
	req := new(ListRequest)

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		rest.RespondWithErrorJSON(w, http.StatusBadRequest, "Invalid request payload for listing files.", err)
		return
	}

	files, err := walkDir(filepath.Join(root, req.SubDir))
	if err != nil {
		rest.RespondWithErrorJSON(w, http.StatusBadRequest, "Error reading directory.", err) // Or InternalServerError depending on expected errors
		return
	}

	if req.OrderBy == "modtime" {
		sort.SliceStable(*files, func(i, j int) bool {
			return (*files)[i].ModTime.After((*files)[j].ModTime)
		})
	}

	w.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(w).Encode(files); err != nil {
		rest.RespondWithErrorJSON(w, http.StatusInternalServerError, "Failed to encode file list response.", err)
	}
}

type DeleteRequest = DirectoryEntry

func DeleteFile(w http.ResponseWriter, r *http.Request) {
	req := new(DeleteRequest)

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		rest.RespondWithErrorJSON(w, http.StatusBadRequest, "Invalid request payload for deleting file.", err)
		return
	}

	if err := os.Remove(req.Path); err != nil {
		rest.RespondWithErrorJSON(w, http.StatusInternalServerError, "Failed to delete file.", err) // Changed to InternalServerError as os.Remove can fail for various reasons
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode("ok")
}

func SendFile(w http.ResponseWriter, r *http.Request) {
	path := chi.URLParam(r, "id")

	if path == "" {
		rest.RespondWithErrorJSON(w, http.StatusBadRequest, "File path is required.", nil)
		return
	}

	path, err := url.QueryUnescape(path)
	if err != nil {
		rest.RespondWithErrorJSON(w, http.StatusBadRequest, "Invalid URL path encoding.", err)
		return
	}

	decoded, err := base64.StdEncoding.DecodeString(path)
	if err != nil {
		rest.RespondWithErrorJSON(w, http.StatusBadRequest, "Invalid base64 encoding in path.", err)
		return
	}

	filename := string(decoded)

	root := config.Instance().DownloadPath

	if strings.Contains(filepath.Dir(filepath.Clean(filename)), filepath.Clean(root)) {
		http.ServeFile(w, r, filename)
		return
	}

	w.WriteHeader(http.StatusUnauthorized)
}

func DownloadFile(w http.ResponseWriter, r *http.Request) {
	path := chi.URLParam(r, "id")

	if path == "" {
		rest.RespondWithErrorJSON(w, http.StatusBadRequest, "File path is required for download.", nil)
		return
	}

	path, err := url.QueryUnescape(path)
	if err != nil {
		rest.RespondWithErrorJSON(w, http.StatusBadRequest, "Invalid URL path encoding for download.", err)
		return
	}

	decoded, err := base64.StdEncoding.DecodeString(path)
	if err != nil {
		rest.RespondWithErrorJSON(w, http.StatusBadRequest, "Invalid base64 encoding in path for download.", err)
		return
	}

	filename := string(decoded)

	root := config.Instance().DownloadPath

	if strings.Contains(filepath.Dir(filepath.Clean(filename)), filepath.Clean(root)) {
		w.Header().Add("Content-Disposition", "inline; filename=\""+filepath.Base(filename)+"\"")
		w.Header().Set("Content-Type", "application/octet-stream")

		fd, err := os.Open(filename)
		if err != nil {
			rest.RespondWithErrorJSON(w, http.StatusInternalServerError, "Failed to open file for download.", err)
			return
		}

		io.Copy(w, fd) // If io.Copy fails, it's hard to send a JSON error as headers might be sent.
		return
	}

	w.WriteHeader(http.StatusUnauthorized)
}

func BulkDownload(mdb *internal.MemoryDB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ps := slices.DeleteFunc(*mdb.All(), func(e internal.ProcessResponse) bool {
			return e.Progress.Status != internal.StatusCompleted
		})

		if len(ps) == 0 {
			return
		}

		zipWriter := zip.NewWriter(w)

		w.Header().Add(
			"Content-Disposition",
			"inline; filename=download-archive-"+time.Now().Format(time.RFC3339)+".zip",
		)
		w.Header().Set("Content-Type", "application/zip")

		for _, p := range ps {
			wr, err := zipWriter.Create(filepath.Base(p.Output.SavedFilePath))
			if err != nil {
				// Difficult to send JSON error here as headers for zip are likely sent. Log and return.
				slog.Error("Failed to create zip entry", "file", p.Output.SavedFilePath, "error", err)
				// We can't use RespondWithErrorJSON effectively if headers are already written.
				return
			}

			fd, err := os.Open(p.Output.SavedFilePath)
			if err != nil {
				slog.Error("Failed to open file for zipping", "file", p.Output.SavedFilePath, "error", err)
				// We can't use RespondWithErrorJSON effectively if headers are already written.
				return
			}

			if _, err := io.Copy(wr, fd); err != nil {
				slog.Error("Failed to copy file to zip", "file", p.Output.SavedFilePath, "error", err)
				// We can't use RespondWithErrorJSON effectively if headers are already written.
				return
			}
			fd.Close() // Close the file descriptor
		}

		if err := zipWriter.Close(); err != nil {
			// Difficult to send JSON error here. Log and return.
			slog.Error("Failed to close zip writer", "error", err)
			// We can't use RespondWithErrorJSON effectively if headers are already written.
			return
		}
	}
}
