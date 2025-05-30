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

type Handler struct { // Using Handler as per prompt and previous file state
	service domain.Service
}

func New(service domain.Service) domain.RestHandler {
	return &Handler{ // Using Handler
		service: service,
	}
}


// List implements domain.RestHandler.
func (h *Handler) List() http.HandlerFunc { // Using Handler
	return func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		w.Header().Set("Content-Type", "application/json")

		query := r.URL.Query()
		var (
			startRowIdParam  = query.Get("id")
			limitParam       = query.Get("limit")
			sortByParam      = query.Get("sort_by")
			// New filter params from query
			filterUploader    = query.Get("filter_uploader")
			filterFormat      = query.Get("filter_format")
			filterMinDuration = query.Get("filter_min_duration")
			filterMaxDuration = query.Get("filter_max_duration")
			searchQueryParam  = query.Get("search_query") // New
		)

		startRowId, err := strconv.Atoi(startRowIdParam)
		if err != nil {
			startRowId = 0
		}

		limit, err := strconv.Atoi(limitParam)
		if err != nil {
			limit = 50 // Default limit
		}

		filters := make(map[string]string)
		if filterUploader != "" {
			filters["uploader"] = filterUploader
		}
		if filterFormat != "" {
			filters["format"] = filterFormat
		}
		if filterMinDuration != "" {
			filters["min_duration"] = filterMinDuration
		}
		if filterMaxDuration != "" {
			filters["max_duration"] = filterMaxDuration
		}
		// Add more filters here as needed

		slog.Info("Archive List Request", 
			"startRowId", startRowId, 
			"limit", limit, 
			"sortBy", sortByParam, 
			"filters", filters,
			"searchQuery", searchQueryParam, // New
		)

		res, err := h.service.List(r.Context(), startRowId, limit, sortByParam, filters, searchQueryParam) // Pass searchQueryParam
		if err != nil {
			slog.Error("Error from archive service List", "error", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		if err := json.NewEncoder(w).Encode(res); err != nil {
			slog.Error("Failed to encode archive list response", "error", err)
			// Consider returning http.Error here as well
			http.Error(w, "Failed to encode response", http.StatusInternalServerError) // Added error response
			return // Added return
		}
	}
}

// ApplyRouter, Archive, SoftDelete, HardDelete, GetCursor methods remain here...
func (h *Handler) ApplyRouter() func(chi.Router) { // Using Handler
	return func(r chi.Router) {
		if config.Instance().RequireAuth {
			r.Use(middlewares.Authenticated)
		}
		if config.Instance().UseOpenId {
			r.Use(openid.Middleware)
		}

		r.Get("/", h.List())
		r.Get("/cursor/{id}", h.GetCursor())
		r.Post("/", h.Archive())
		r.Delete("/soft/{id}", h.SoftDelete())
		r.Delete("/hard/{id}", h.HardDelete())
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

		id := chi.URLParam(r, "id") 
		if id == "" { 
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
