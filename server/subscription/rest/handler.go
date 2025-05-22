package rest

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/marcopiovanello/yt-dlp-web-ui/v3/server/config"
	middlewares "github.com/marcopiovanello/yt-dlp-web-ui/v3/server/middleware"
	"github.com/marcopiovanello/yt-dlp-web-ui/v3/server/openid"
	"github.com/marcopiovanello/yt-dlp-web-ui/v3/server/subscription/domain"
	"log/slog" // Added for logging
)

type RestHandler struct {
	svc domain.Service
}

func New(svc domain.Service) domain.RestHandler { // Ensure New returns domain.RestHandler
	return &RestHandler{
		svc: svc,
	}
}

// ApplyRouter implements domain.RestHandler.
func (h *RestHandler) ApplyRouter() func(chi.Router) {
	return func(r chi.Router) {
		if config.Instance().RequireAuth {
			r.Use(middlewares.Authenticated)
		}
		if config.Instance().UseOpenId {
			r.Use(openid.Middleware)
		}

		r.Delete("/{id}", h.Delete())
		r.Get("/cursor", h.GetCursor())
		r.Get("/", h.List())
		r.Post("/", h.Submit())
		r.Patch("/", h.UpdateByExample())
		r.Get("/{id}/videos", h.GetChannelVideos()) // New route
	}
}

// GetChannelVideos handles fetching channel videos metadata for a subscription.
func (h *RestHandler) GetChannelVideos() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		w.Header().Set("Content-Type", "application/json")

		subscriptionID := chi.URLParam(r, "id")
		if subscriptionID == "" {
			slog.Error("Subscription ID is missing in path")
			http.Error(w, "Subscription ID is required", http.StatusBadRequest)
			return
		}

		slog.Info("Handler: GetChannelVideos called", "subscriptionID", subscriptionID)

		channelDump, err := h.svc.GetChannelVideos(r.Context(), subscriptionID)
		if err != nil {
			slog.Error("Error from GetChannelVideos service", "subscriptionID", subscriptionID, "error", err)
			// Distinguish between "not found" and other errors if possible
			// For now, generic internal server error.
			http.Error(w, "Failed to get channel videos: "+err.Error(), http.StatusInternalServerError)
			return
		}

		if err := json.NewEncoder(w).Encode(channelDump); err != nil {
			slog.Error("Failed to encode channel videos response", "subscriptionID", subscriptionID, "error", err)
			http.Error(w, "Failed to encode response", http.StatusInternalServerError)
			return
		}
	}
}

// Delete implements domain.RestHandler.
func (h *RestHandler) Delete() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		w.Header().Set("Content-Type", "application/json")

		id := chi.URLParam(r, "id")

		err := h.svc.Delete(r.Context(), id)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		if err := json.NewEncoder(w).Encode("ok"); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
}

// GetCursor implements domain.RestHandler.
func (h *RestHandler) GetCursor() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		w.Header().Set("Content-Type", "application/json")

		id := chi.URLParam(r, "id")

		cursorId, err := h.svc.GetCursor(r.Context(), id)
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

// List implements domain.RestHandler.
func (h *RestHandler) List() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		w.Header().Set("Content-Type", "application/json")

		var (
			startParam = r.URL.Query().Get("id")
			LimitParam = r.URL.Query().Get("limit")
		)

		start, err := strconv.Atoi(startParam)
		if err != nil {
			start = 0
		}

		limit, err := strconv.Atoi(LimitParam)
		if err != nil {
			limit = 50
		}

		res, err := h.svc.List(r.Context(), int64(start), limit)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		if err := json.NewEncoder(w).Encode(res); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
}

// Submit implements domain.RestHandler.
func (h *RestHandler) Submit() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		w.Header().Set("Content-Type", "application/json")

		var req domain.Subscription

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		res, err := h.svc.Submit(r.Context(), &req)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		if err := json.NewEncoder(w).Encode(res); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
}

// UpdateByExample implements domain.RestHandler.
func (h *RestHandler) UpdateByExample() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		w.Header().Set("Content-Type", "application/json")

		var req domain.Subscription

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		if err := h.svc.UpdateByExample(r.Context(), &req); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		if err := json.NewEncoder(w).Encode(req); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
}

// Ensure New is here or that the one at the top is sufficient.
// The diff has placed New() higher up, which is fine.
