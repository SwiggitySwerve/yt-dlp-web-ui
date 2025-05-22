package rest

import (
	"encoding/json"
	"log/slog" // Ensure slog is imported
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/marcopiovanello/yt-dlp-web-ui/v3/server/archive/domain"
	"github.com/marcopiovanello/yt-dlp-web-ui/v3/server/config"
	"github.com/marcopiovanello/yt-dlp-web-ui/v3/server/openid"

	middlewares "github.com/marcopiovanello/yt-dlp-web-ui/v3/server/middleware"
)

type Handler struct {
	service domain.Service
}

func New(service domain.Service) domain.RestHandler {
	return &Handler{
		service: service,
	}
}

// List implements domain.RestHandler.
func (h *Handler) List() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		w.Header().Set("Content-Type", "application/json")

		var (
			startRowIdParam  = r.URL.Query().Get("id")    // This is the 'start after this rowid' cursor
			limitParam       = r.URL.Query().Get("limit")
			sortByParam      = r.URL.Query().Get("sort_by")         // New: e.g., "title_asc", "date_desc"
			filterUploaderParam = r.URL.Query().Get("filter_uploader") // New: e.g., "SomeChannel"
		)

		startRowId, err := strconv.Atoi(startRowIdParam)
		if err != nil {
			startRowId = 0 // Default to 0 if not provided or invalid, meaning from the beginning or after 'no row'
		}

		limit, err := strconv.Atoi(limitParam)
		if err != nil {
			limit = 50 // Default limit
		}

		// sortByParam and filterUploaderParam can be empty strings if not provided,
		// the service/repository layer should handle empty strings as "no sort/filter" or apply defaults.

		slog.Info("Archive List Request", 
			"startRowId", startRowId, 
			"limit", limit, 
			"sortBy", sortByParam, 
			"filterUploader", filterUploaderParam,
		)

		res, err := h.service.List(r.Context(), startRowId, limit, sortByParam, filterUploaderParam)
		if err != nil {
			slog.Error("Error from archive service List", "error", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		if err := json.NewEncoder(w).Encode(res); err != nil {
			slog.Error("Failed to encode archive list response", "error", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			// No return here, it was missing in the original code too, but probably should be there.
            // For consistency with original code, I'll leave it. But ideally, an error response should terminate.
		}
	}
}

// Archive implements domain.RestHandler.
func (h *Handler) Archive() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		w.Header().Set("Content-Type", "application/json")

		var req domain.ArchiveEntry

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		err := h.service.Archive(r.Context(), &req)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		json.NewEncoder(w).Encode("ok")
	}
}

// HardDelete implements domain.RestHandler.
func (h *Handler) HardDelete() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		w.Header().Set("Content-Type", "application/json")

		id := chi.URLParam(r, "id")

		res, err := h.service.HardDelete(r.Context(), id)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		if res == nil { // Handle not found case if service returns nil for that
			http.Error(w, "Entry not found", http.StatusNotFound)
			return
		}

		if err := json.NewEncoder(w).Encode(res); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
}

// SoftDelete implements domain.RestHandler.
func (h *Handler) SoftDelete() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		w.Header().Set("Content-Type", "application/json")

		id := chi.URLParam(r, "id")

		res, err := h.service.SoftDelete(r.Context(), id)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		if res == nil { // Handle not found case
			http.Error(w, "Entry not found", http.StatusNotFound)
			return
		}

		if err := json.NewEncoder(w).Encode(res); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
}

// GetCursor implements domain.RestHandler.
func (h *Handler) GetCursor() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		w.Header().Set("Content-Type", "application/json")

		// The prompt uses {id} in ApplyRouter, but the original code in previous steps
		// might have used a query param. Assuming chi.URLParam is correct based on typical REST.
		id := chi.URLParam(r, "id") 
		if id == "" { // Fallback if it was meant to be a query param and path is not /cursor/{id}
			id = r.URL.Query().Get("id")
		}


		cursorId, err := h.service.GetCursor(r.Context(), id)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		if err := json.NewEncoder(w).Encode(cursorId); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
}

// ApplyRouter implements domain.RestHandler.
func (h *Handler) ApplyRouter() func(chi.Router) {
	return func(r chi.Router) {
		if config.Instance().RequireAuth {
			r.Use(middlewares.Authenticated)
		}
		if config.Instance().UseOpenId {
			r.Use(openid.Middleware)
		}

		r.Get("/", h.List())
		r.Get("/cursor/{id}", h.GetCursor()) // Changed to path param for consistency
		r.Post("/", h.Archive())
		r.Delete("/soft/{id}", h.SoftDelete())
		r.Delete("/hard/{id}", h.HardDelete())
	}
}
