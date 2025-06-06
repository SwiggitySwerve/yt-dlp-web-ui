package rest

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/marcopiovanello/yt-dlp-web-ui/v3/server/internal"
)

type Handler struct {
	service *Service
}

/*
	REST version of the JSON-RPC interface
*/

func (h *Handler) Exec() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()

		w.Header().Set("Content-Type", "application/json")

		var req internal.DownloadRequest

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			RespondWithErrorJSON(w, http.StatusBadRequest, "Invalid request payload for Exec.", err) // Changed from StatusInternalServerError
			return // Added return
		}

		id, err := h.service.Exec(req)
		if err != nil {
			RespondWithErrorJSON(w, http.StatusInternalServerError, "Failed to execute download.", err)
			return
		}

		if err := json.NewEncoder(w).Encode(id); err != nil {
			RespondWithErrorJSON(w, http.StatusInternalServerError, "Failed to encode Exec response.", err)
			return
		}
	}
}

func (h *Handler) ExecPlaylist() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()

		w.Header().Set("Content-Type", "application/json")

		var req internal.DownloadRequest

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			RespondWithErrorJSON(w, http.StatusBadRequest, "Invalid request payload for ExecPlaylist.", err) // Changed from StatusInternalServerError
			return // Added return
		}

		err := h.service.ExecPlaylist(req)
		if err != nil {
			RespondWithErrorJSON(w, http.StatusInternalServerError, "Failed to execute playlist download.", err)
			return
		}

		if err := json.NewEncoder(w).Encode("ok"); err != nil {
			RespondWithErrorJSON(w, http.StatusInternalServerError, "Failed to encode ExecPlaylist response.", err)
			return
		}
	}
}

func (h *Handler) ExecLivestream() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()

		w.Header().Set("Content-Type", "application/json")

		var req internal.DownloadRequest

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			RespondWithErrorJSON(w, http.StatusBadRequest, "Invalid request payload for ExecLivestream.", err)
			return
		}

		h.service.ExecLivestream(req) // Assuming ExecLivestream logs its own errors if critical, or doesn't return one to handler

		if err := json.NewEncoder(w).Encode("ok"); err != nil {
			RespondWithErrorJSON(w, http.StatusInternalServerError, "Failed to encode ExecLivestream response.", err)
			return
		}
	}
}

func (h *Handler) Running() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()

		w.Header().Set("Content-Type", "application/json")

		res, err := h.service.Running(r.Context())
		if err != nil {
			RespondWithErrorJSON(w, http.StatusInternalServerError, "Failed to get running processes.", err)
			return
		}

		if err := json.NewEncoder(w).Encode(res); err != nil {
			RespondWithErrorJSON(w, http.StatusInternalServerError, "Failed to encode running processes response.", err)
			return
		}
	}
}

func (h *Handler) GetCookies() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		cookies, err := h.service.GetCookies(r.Context())
		if err != nil {
			RespondWithErrorJSON(w, http.StatusInternalServerError, "Failed to get cookies.", err)
			return
		}

		res := &internal.SetCookiesRequest{
			Cookies: string(cookies),
		}

		if err := json.NewEncoder(w).Encode(res); err != nil {
			RespondWithErrorJSON(w, http.StatusInternalServerError, "Failed to encode get cookies response.", err)
			return
		}
	}
}

func (h *Handler) SetCookies() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()

		w.Header().Set("Content-Type", "application/json")

		req := new(internal.SetCookiesRequest)

		if err := json.NewDecoder(r.Body).Decode(req); err != nil {
			RespondWithErrorJSON(w, http.StatusBadRequest, "Invalid request payload for SetCookies.", err)
			return
		}

		if err := h.service.SetCookies(r.Context(), req.Cookies); err != nil {
			RespondWithErrorJSON(w, http.StatusInternalServerError, "Failed to set cookies.", err)
			return
		}

		if err := json.NewEncoder(w).Encode("ok"); err != nil {
			RespondWithErrorJSON(w, http.StatusInternalServerError, "Failed to encode SetCookies response.", err)
			return
		}
	}
}

func (h *Handler) DeleteCookies() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		if err := h.service.SetCookies(r.Context(), ""); err != nil {
			RespondWithErrorJSON(w, http.StatusInternalServerError, "Failed to delete cookies.", err)
			return
		}

		if err := json.NewEncoder(w).Encode("ok"); err != nil {
			RespondWithErrorJSON(w, http.StatusInternalServerError, "Failed to encode DeleteCookies response.", err)
			return
		}
	}
}

func (h *Handler) AddTemplate() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()

		w.Header().Set("Content-Type", "application/json")

		req := new(internal.CustomTemplate)

		if err := json.NewDecoder(r.Body).Decode(req); err != nil {
			RespondWithErrorJSON(w, http.StatusBadRequest, "Invalid request payload for AddTemplate.", err)
			return
		}

		if req.Name == "" || req.Content == "" {
			RespondWithErrorJSON(w, http.StatusBadRequest, "Invalid template: name and content are required.", nil)
			return
		}

		if err := h.service.SaveTemplate(r.Context(), req); err != nil {
			RespondWithErrorJSON(w, http.StatusInternalServerError, "Failed to save template.", err)
			return
		}

		if err := json.NewEncoder(w).Encode("ok"); err != nil {
			RespondWithErrorJSON(w, http.StatusInternalServerError, "Failed to encode AddTemplate response.", err)
			return
		}
	}
}

func (h *Handler) GetTemplates() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()

		w.Header().Set("Content-Type", "application/json")

		templates, err := h.service.GetTemplates(r.Context())
		if err != nil {
			RespondWithErrorJSON(w, http.StatusInternalServerError, "Failed to get templates.", err)
			return
		}

		err = json.NewEncoder(w).Encode(templates)
		if err != nil {
			RespondWithErrorJSON(w, http.StatusInternalServerError, "Failed to encode GetTemplates response.", err)
		}
	}
}

func (h *Handler) UpdateTemplate() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()

		w.Header().Set("Content-Type", "application/json")

		req := &internal.CustomTemplate{}

		if err := json.NewDecoder(r.Body).Decode(req); err != nil {
			RespondWithErrorJSON(w, http.StatusBadRequest, "Invalid request payload for UpdateTemplate.", err) // Changed to StatusBadRequest
			return
		}

		res, err := h.service.UpdateTemplate(r.Context(), req)

		if err != nil {
			RespondWithErrorJSON(w, http.StatusInternalServerError, "Failed to update template.", err)
			return
		}

		if err := json.NewEncoder(w).Encode(res); err != nil {
			RespondWithErrorJSON(w, http.StatusInternalServerError, "Failed to encode UpdateTemplate response.", err)
			return
		}
	}
}

func (h *Handler) DeleteTemplate() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()

		w.Header().Set("Content-Type", "application/json")

		id := chi.URLParam(r, "id")

		if err := h.service.DeleteTemplate(r.Context(), id); err != nil {
			RespondWithErrorJSON(w, http.StatusInternalServerError, "Failed to delete template.", err)
			return
		}

		if err := json.NewEncoder(w).Encode("ok"); err != nil {
			RespondWithErrorJSON(w, http.StatusInternalServerError, "Failed to encode DeleteTemplate response.", err)
			return
		}
	}
}

func (h *Handler) GetVersion() http.HandlerFunc {
	type Response struct {
		RPCVersion   string `json:"rpcVersion"`
		YtdlpVersion string `json:"ytdlpVersion"`
	}

	return func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()

		w.Header().Set("Content-Type", "application/json")

		rpcVersion, ytdlpVersion, err := h.service.GetVersion(r.Context())
		if err != nil {
			RespondWithErrorJSON(w, http.StatusInternalServerError, "Failed to get version information.", err)
			return
		}

		res := Response{
			RPCVersion:   rpcVersion,
			YtdlpVersion: ytdlpVersion,
		}

		if err := json.NewEncoder(w).Encode(res); err != nil {
			RespondWithErrorJSON(w, http.StatusInternalServerError, "Failed to encode version information response.", err)
			return
		}
	}
}
