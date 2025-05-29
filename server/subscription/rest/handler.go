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

// --- Implementations for SubscriptionVideoUpdate handlers ---

func (h *RestHandler) ListUpdates() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		w.Header().Set("Content-Type", "application/json")

		query := r.URL.Query()
		limitParam := query.Get("limit")
		offsetParam := query.Get("offset")
		subIDsParam := query.Get("subscription_ids")

		limit, err := strconv.Atoi(limitParam)
		if err != nil || limit <= 0 {
			limit = 20 // Default limit
		}
		offset, err := strconv.Atoi(offsetParam)
		if err != nil || offset < 0 {
			offset = 0 // Default offset
		}

		var parsedSubscriptionIDs []string
		if subIDsParam != "" {
			parsedSubscriptionIDs = strings.Split(subIDsParam, ",")
		}

		slog.Info("Handler: ListUpdates called", "limit", limit, "offset", offset, "subscriptionIDs", parsedSubscriptionIDs)
		updates, err := h.svc.ListUnseenUpdates(r.Context(), limit, offset, parsedSubscriptionIDs)
		if err != nil {
			slog.Error("Error from ListUnseenUpdates service", "error", err)
			http.Error(w, "Failed to list updates: "+err.Error(), http.StatusInternalServerError)
			return
		}

		if err := json.NewEncoder(w).Encode(updates); err != nil {
			slog.Error("Failed to encode updates list response", "error", err)
			http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		}
	}
}

func (h *RestHandler) GetUnseenUpdatesCount() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		w.Header().Set("Content-Type", "application/json")

		subIDsParam := r.URL.Query().Get("subscription_ids")
		var parsedSubscriptionIDs []string
		if subIDsParam != "" {
			parsedSubscriptionIDs = strings.Split(subIDsParam, ",")
		}
		
		slog.Info("Handler: GetUnseenUpdatesCount called", "subscriptionIDs", parsedSubscriptionIDs)
		count, err := h.svc.GetUnseenUpdatesCount(r.Context(), parsedSubscriptionIDs)
		if err != nil {
			slog.Error("Error from GetUnseenUpdatesCount service", "error", err)
			http.Error(w, "Failed to get unseen updates count: "+err.Error(), http.StatusInternalServerError)
			return
		}

		if err := json.NewEncoder(w).Encode(map[string]int{"count": count}); err != nil {
			slog.Error("Failed to encode unseen updates count response", "error", err)
			http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		}
	}
}

func (h *RestHandler) MarkUpdateSeen() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		w.Header().Set("Content-Type", "application/json")

		updateID := chi.URLParam(r, "updateID")
		if updateID == "" {
			slog.Error("Update ID is missing in path for MarkUpdateSeen")
			http.Error(w, "Update ID is required", http.StatusBadRequest)
			return
		}

		var reqBody struct {
			Seen bool `json:"seen"`
		}
		if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
			slog.Error("Failed to decode request body for MarkUpdateSeen", "error", err)
			http.Error(w, "Invalid request body: "+err.Error(), http.StatusBadRequest)
			return
		}
		
		slog.Info("Handler: MarkUpdateSeen called", "updateID", updateID, "seen", reqBody.Seen)
		err := h.svc.MarkUpdateAsSeen(r.Context(), updateID, reqBody.Seen)
		if err != nil {
			slog.Error("Error from MarkUpdateAsSeen service", "updateID", updateID, "error", err)
			http.Error(w, "Failed to mark update as seen: "+err.Error(), http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"message": "Update marked as seen successfully"})
	}
}

func (h *RestHandler) MarkAllUpdatesSeen() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		w.Header().Set("Content-Type", "application/json")

		subIDsParam := r.URL.Query().Get("subscription_ids")
		var parsedSubscriptionIDs []string
		if subIDsParam != "" {
			parsedSubscriptionIDs = strings.Split(subIDsParam, ",")
		}

		var reqBody struct {
			Seen bool `json:"seen"`
		}
		if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
			slog.Error("Failed to decode request body for MarkAllUpdatesSeen", "error", err)
			http.Error(w, "Invalid request body: "+err.Error(), http.StatusBadRequest)
			return
		}
		
		slog.Info("Handler: MarkAllUpdatesSeen called", "subscriptionIDs", parsedSubscriptionIDs, "seen", reqBody.Seen)
		affectedCount, err := h.svc.MarkAllUpdatesAsSeen(r.Context(), parsedSubscriptionIDs, reqBody.Seen)
		if err != nil {
			slog.Error("Error from MarkAllUpdatesAsSeen service", "error", err)
			http.Error(w, "Failed to mark all updates as seen: "+err.Error(), http.StatusInternalServerError)
			return
		}
		json.NewEncoder(w).Encode(map[string]interface{}{"message": "All updates marked as seen successfully", "affected_count": affectedCount})
	}
}

func (h *RestHandler) DownloadUpdate() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		w.Header().Set("Content-Type", "application/json")

		updateID := chi.URLParam(r, "updateID")
		if updateID == "" {
			slog.Error("Update ID is missing in path for DownloadUpdate")
			http.Error(w, "Update ID is required", http.StatusBadRequest)
			return
		}
		
		slog.Info("Handler: DownloadUpdate called", "updateID", updateID)

		// Assuming h.svc has GetSubscriptionUpdate method (added to domain.Service in Turn 54)
		videoUpdate, err := h.svc.GetSubscriptionUpdate(r.Context(), updateID)
		if err != nil {
			slog.Error("Error from GetSubscriptionUpdate service", "updateID", updateID, "error", err)
			http.Error(w, "Failed to get video update details: "+err.Error(), http.StatusInternalServerError)
			return
		}
		if videoUpdate == nil {
			slog.Error("Video update not found", "updateID", updateID)
			http.Error(w, "Video update not found", http.StatusNotFound)
			return
		}

		p := &internal.Process{
			Url:        videoUpdate.VideoURL,
			Params:     []string{}, // Placeholder: Consider how to get relevant params
			AutoRemove: false,      
		}
		h.memDB.Set(p)  // Generate ID for the process
		h.mq.Publish(p) // Queue for download

		statusUpdateErr := h.svc.UpdateSubscriptionUpdateStatus(r.Context(), updateID, "queued_for_download")
		if statusUpdateErr != nil {
			slog.Error("Failed to update video update status after queuing for download", "updateID", updateID, "error", statusUpdateErr)
		}

		w.WriteHeader(http.StatusAccepted)
		json.NewEncoder(w).Encode(map[string]string{"message": "Video queued for download", "process_id": p.Id})
	}
}

func (h *RestHandler) DeleteUpdate() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		w.Header().Set("Content-Type", "application/json")

		updateID := chi.URLParam(r, "updateID")
		if updateID == "" {
			slog.Error("Update ID is missing in path for DeleteUpdate")
			http.Error(w, "Update ID is required", http.StatusBadRequest)
			return
		}
		
		slog.Info("Handler: DeleteUpdate called", "updateID", updateID)
		err := h.svc.DeleteSubscriptionUpdate(r.Context(), updateID)
		if err != nil {
			slog.Error("Error from DeleteSubscriptionUpdate service", "updateID", updateID, "error", err)
			http.Error(w, "Failed to delete update: "+err.Error(), http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"message": "Update deleted successfully"})
	}
}
