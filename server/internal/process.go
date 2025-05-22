package internal

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"path/filepath" 
	"regexp"
	"slices" 
	"syscall"

	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/marcopiovanello/yt-dlp-web-ui/v3/server/archiver"
	"github.com/marcopiovanello/yt-dlp-web-ui/v3/server/common"
	"github.com/marcopiovanello/yt-dlp-web-ui/v3/server/config"
)

const downloadTemplate = `download:
{
	"eta":%(progress.eta)s,
	"percentage":"%(progress._percent_str)s",
	"speed":%(progress.speed)s
}`

const postprocessTemplate = `postprocess:
{
	"filepath":"%(info.filepath)s"
}
`

const (
	StatusPending = iota
	StatusDownloading
	StatusCompleted
	StatusErrored
)

type Process struct {
	Id         string
	Url        string
	Livestream bool
	AutoRemove bool
	Params     []string
	Info       common.DownloadInfo
	Progress   DownloadProgress
	Output     DownloadOutput 
	proc       *os.Process
	PreferredFormats []string 
	PreferredQualities []string 
}

type DownloadOutput struct {
	Path          string
	Filename      string
	SavedFilePath string
	ChannelFolder string 
}

func (p *Process) Start() {
	p.Params = slices.DeleteFunc(p.Params, func(e string) bool {
		match, _ := regexp.MatchString(`(\$\{)|(\&\&)`, e)
		return match
	})
	p.Params = slices.DeleteFunc(p.Params, func(e string) bool {
		return e == ""
	})

	if p.Output.Path == "" {
        p.Output.Path = config.Instance().DownloadPath
    }
	if p.Output.Filename == "" {
        p.Output.Filename = "%(title)s.%(ext)s"
    }
	buildFilename(&p.Output) 

	templateReplacer := strings.NewReplacer("\n", "", "\t", "", " ", "")
	
	// Start assembling final command parameters
	currentParams := []string{
		strings.Split(p.Url, "?list")[0], //no playlist
		"--newline",
		"--no-colors",
		"--no-playlist", 
		"--progress-template",
		templateReplacer.Replace(downloadTemplate),
		"--progress-template",
		templateReplacer.Replace(postprocessTemplate),
	}

	// Logic to Construct Format String
	var formatString string
	if (len(p.PreferredQualities) > 0 || len(p.PreferredFormats) > 0) { // Modified condition
		slog.Info("Constructing format string from preferences", "qualities", p.PreferredQualities, "formats", p.PreferredFormats)
		var formatArgs []string

        // Prioritize qualities, then formats, then best
        // Example: 1080p60.mp4/1080p.mp4/best.mp4/1080p60/1080p/best
        // Simplified: (q1.f1/q1.f2/q1)/(q2.f1/q2.f2/q2)/best.f1/best.f2/best
        
        // Handle qualities
        if len(p.PreferredQualities) > 0 {
            for _, quality := range p.PreferredQualities {
                qualitySelector := quality // e.g., "1080p", "720p", "best"
                if quality != "best" && strings.HasSuffix(quality, "p") { // "1080p" -> "height<=1080"
                    qualitySelector = fmt.Sprintf("bv*[height<=%s]", strings.TrimSuffix(quality, "p"))
                } else if quality == "best" {
                    qualitySelector = "bv*" // bestvideo shorthand
                }
                // If specific formats are also preferred, combine them
                if len(p.PreferredFormats) > 0 {
                    for _, formatExt := range p.PreferredFormats {
                        formatArgs = append(formatArgs, fmt.Sprintf("%s[ext=%s]+ba[ext=%s]/%s[ext=%s]", qualitySelector, formatExt, formatExt, qualitySelector, formatExt))
                    }
                }
                // Add the quality selector itself (e.g., "bv*[height<=1080]" or "bv*")
                formatArgs = append(formatArgs, qualitySelector)
            }
        }

        // Handle formats if no qualities specified, or as a lower priority fallback
        if len(p.PreferredFormats) > 0 {
             for _, formatExt := range p.PreferredFormats {
                formatArgs = append(formatArgs, fmt.Sprintf("b[ext=%s]", formatExt)) // best with specific extension
             }
        }
        
        formatArgs = append(formatArgs, "best") // Overall fallback

		formatString = strings.Join(formatArgs, "/")
		slog.Info("Constructed yt-dlp format string", "format", formatString)
	}

	// Check if user p.Params already contains -f or --format.
	hasUserFormat := false
	for i, param := range p.Params {
		if param == "-f" || param == "--format" {
			hasUserFormat = true
			// If user provides format, and it's not just the flag but also a value
			if i+1 < len(p.Params) {
				slog.Info("User provided format string in raw params", "userFormat", p.Params[i+1])
			}
			break
		}
	}

	if formatString != "" && !hasUserFormat {
		currentParams = append(currentParams, "-f", formatString)
		slog.Info("Using preference-generated format string", "format", formatString)
	} else if hasUserFormat {
        slog.Info("Using user-provided format string from raw parameters.")
    }
    // If neither, yt-dlp uses its default "best".

	// Append user's raw parameters (p.Params)
    // These are added after our potential format string. If user also included -f, yt-dlp behavior
    // (usually last -f wins, or they can be additive if complex) applies.
    // Our check for `hasUserFormat` above is mostly for logging/decision if we *should* add ours.
	currentParams = append(currentParams, p.Params...)
	
	// Output path logic
	var fullOutputPath string
	if p.Output.ChannelFolder != "" {
		fullOutputPath = filepath.Join(p.Output.Path, p.Output.ChannelFolder, p.Output.Filename)
	} else {
		fullOutputPath = filepath.Join(p.Output.Path, p.Output.Filename)
	}

	// Add -o unless user provided -P or --paths in their original p.Params
    userProvidedOutputFlag := false
    for _, param := range p.Params { // Check original p.Params
        if param == "-P" || param == "--paths" || param == "-o" {
            userProvidedOutputFlag = true
            break
        }
    }
	if !userProvidedOutputFlag {
		currentParams = append(currentParams, "-o", fullOutputPath)
	}

	slog.Info("requesting download", slog.String("url", p.Url), slog.Any("params", currentParams))
	cmd := exec.Command(config.Instance().DownloaderPath, currentParams...)
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		slog.Error("failed to get a stdout pipe", slog.Any("err", err))
		panic(err)
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		slog.Error("failed to get a stderr pipe", slog.Any("err", err))
		panic(err)
	}
	if err := cmd.Start(); err != nil {
		slog.Error("failed to start yt-dlp process", slog.Any("err", err))
		panic(err)
	}
	p.proc = cmd.Process
	ctx, cancel := context.WithCancel(context.Background())
	defer func() {
		stdout.Close()
		p.Complete()
		cancel()
	}()
	logs := make(chan []byte)
	go produceLogs(stdout, logs)
	go p.consumeLogs(ctx, logs)
	go p.detectYtDlpErrors(stderr)
	cmd.Wait()
}

// ... (rest of the file remains the same as per previous state) ...

func produceLogs(r io.Reader, logs chan<- []byte) {
	go func() {
		scanner := bufio.NewScanner(r)
		for scanner.Scan() {
			logs <- scanner.Bytes()
		}
	}()
}

func (p *Process) consumeLogs(ctx context.Context, logs <-chan []byte) {
	for {
		select {
		case <-ctx.Done():
			slog.Info("detaching from yt-dlp stdout", slog.String("id", p.getShortId()), slog.String("url", p.Url))
			return
		case entry := <-logs:
			p.parseLogEntry(entry)
		}
	}
}

func (p *Process) parseLogEntry(entry []byte) {
	var progress ProgressTemplate
	var postprocess PostprocessTemplate
	if err := json.Unmarshal(entry, &progress); err == nil {
		p.Progress = DownloadProgress{
			Status:     StatusDownloading,
			Percentage: progress.Percentage,
			Speed:      progress.Speed,
			ETA:        progress.Eta,
		}
		slog.Info("progress", slog.String("id", p.getShortId()), slog.String("url", p.Url), slog.String("percentage", progress.Percentage))
	}
	if err := json.Unmarshal(entry, &postprocess); err == nil {
		p.Output.SavedFilePath = postprocess.FilePath
	}
}

func (p *Process) detectYtDlpErrors(r io.Reader) {
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		slog.Error("yt-dlp process error", slog.String("id", p.getShortId()), slog.String("url", p.Url), slog.String("err", scanner.Text()))
	}
}

func (p *Process) Complete() {
	if p.Progress.Percentage == "" && p.Progress.Speed == 0 {
		var serializedMetadata bytes.Buffer
		json.NewEncoder(&serializedMetadata).Encode(p.Info)
		archiver.Publish(&archiver.Message{
			Id:        p.Id,
			Path:      p.Output.SavedFilePath,
			Title:     p.Info.Title,
			Thumbnail: p.Info.Thumbnail,
			Source:    p.Url,
			Metadata:  serializedMetadata.String(),
			CreatedAt: p.Info.CreatedAt,
		})
	}
	p.Progress = DownloadProgress{
		Status:     StatusCompleted,
		Percentage: "-1",
		Speed:      0,
		ETA:        0,
	}
	if p.Output.SavedFilePath == "" {
		p.GetFileName(&p.Output)
	}
	slog.Info("finished", slog.String("id", p.getShortId()), slog.String("url", p.Url))
	memDbEvents <- p
}

func (p *Process) Kill() error {
	defer func() { p.Progress.Status = StatusCompleted }()
	if p.proc == nil {
		return errors.New("*os.Process not set")
	}
	pgid, err := syscall.Getpgid(p.proc.Pid)
	if err != nil {
		return err
	}
	if err := syscall.Kill(-pgid, syscall.SIGTERM); err != nil {
		return err
	}
	return nil
}

func (p *Process) GetFileName(o *DownloadOutput) error {
	var outputPathArgs []string
	if o.ChannelFolder != "" {
		outputPathArgs = append(outputPathArgs, o.Path, o.ChannelFolder, o.Filename)
	} else {
		outputPathArgs = append(outputPathArgs, o.Path, o.Filename)
	}
	fullOutputTemplate := filepath.Join(outputPathArgs...)
	cmd := exec.Command(
		config.Instance().DownloaderPath,
		"--print", "filename",
		"-o", fullOutputTemplate, 
		p.Url,
	)
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	out, err := cmd.Output()
	if err != nil {
		return err
	}
	p.Output.SavedFilePath = strings.Trim(string(out), "\n")
	return nil
}

func (p *Process) SetPending() {
	p.Info = common.DownloadInfo{
		URL:       p.Url,
		Title:     p.Url,
		CreatedAt: time.Now(),
	}
	p.Progress.Status = StatusPending
}

func (p *Process) SetMetadata() error {
	cmd := exec.Command(config.Instance().DownloaderPath, p.Url, "-J")
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		slog.Error("failed to connect to stdout", slog.String("id", p.getShortId()), slog.String("url", p.Url), slog.String("err", err.Error()))
		return err
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		slog.Error("failed to connect to stderr", slog.String("id", p.getShortId()), slog.String("url", p.Url), slog.String("err", err.Error()))
		return err
	}
	info := common.DownloadInfo{URL: p.Url, CreatedAt: time.Now()}
	if err := cmd.Start(); err != nil {
		return err
	}
	var bufferedStderr bytes.Buffer
	go func() { io.Copy(&bufferedStderr, stderr) }()
	slog.Info("retrieving metadata", slog.String("id", p.getShortId()), slog.String("url", p.Url))
	if err := json.NewDecoder(stdout).Decode(&info); err != nil {
		return err
	}
	p.Info = info
	p.Progress.Status = StatusPending
	if err := cmd.Wait(); err != nil {
		return errors.New(bufferedStderr.String())
	}
	return nil
}

func (p *Process) getShortId() string { return strings.Split(p.Id, "-")[0] }

func buildFilename(o *DownloadOutput) {
	if o.Filename != "" && !strings.Contains(o.Filename, "%(ext)s") {
         o.Filename += ".%(ext)s"
    }
	o.Filename = strings.Replace(o.Filename, ".%(ext)s.%(ext)s", ".%(ext)s", 1)
}

type ProgressTemplate struct {
	ETA        float64 `json:"eta"`
	Percentage string  `json:"percentage"`
	Speed      float64 `json:"speed"`
}

type PostprocessTemplate struct {
	FilePath string `json:"filepath"`
}
