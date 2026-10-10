package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"time"

	"pulsewatch-api/internal/middleware"
	"pulsewatch-api/internal/models"
	"pulsewatch-api/internal/repository"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type MonitorHandler struct {
	repo *repository.MonitorRepository
}

func NewMonitorHandler(repo *repository.MonitorRepository) *MonitorHandler {
	return &MonitorHandler{repo: repo}
}

func (h *MonitorHandler) CreateMonitor(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		JSONError(w, http.StatusUnauthorized, "unauthorized", "authentication required")
		return
	}

	var req CreateMonitorRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		JSONError(w, http.StatusBadRequest, "bad_request", "invalid JSON payload")
		return
	}

	if err := req.Validate(); err != nil {
		JSONError(w, http.StatusUnprocessableEntity, "validation_error", err.Error())
		return
	}

	sslEnabled := true
	if req.SSLCheckEnabled != nil {
		sslEnabled = *req.SSLCheckEnabled
	}

	monitor := models.Monitor{
		ID:                  uuid.New(),
		UserID:              userID,
		Name:                req.Name,
		URL:                 req.URL,
		Method:              req.Method,
		IntervalSeconds:     req.IntervalSeconds,
		TimeoutSeconds:      req.TimeoutSeconds,
		ExpectedStatusCode:  req.ExpectedStatusCode,
		Headers:             req.Headers,
		Body:                req.Body,
		Status:              "PENDING",
		IsActive:            true,
		SSLCheckEnabled:     sslEnabled,
		FailureThreshold:    req.FailureThreshold,
		RecoveryThreshold:   req.RecoveryThreshold,
		NextCheckAt:         time.Now().UTC(),
	}

	if err := h.repo.Create(r.Context(), &monitor, req.AlertChannelIDs); err != nil {
		JSONError(w, http.StatusInternalServerError, "internal_error", "failed to create monitor")
		return
	}

	JSON(w, http.StatusCreated, monitor)
}

func (h *MonitorHandler) ListMonitors(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		JSONError(w, http.StatusUnauthorized, "unauthorized", "authentication required")
		return
	}

	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page < 1 {
		page = 1
	}
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit < 1 || limit > 100 {
		limit = 20
	}
	status := r.URL.Query().Get("status")

	monitors, total, err := h.repo.ListByUserID(r.Context(), userID, page, limit, status)
	if err != nil {
		JSONError(w, http.StatusInternalServerError, "internal_error", "failed to fetch monitors")
		return
	}

	if monitors == nil {
		monitors = []models.Monitor{}
	}

	totalPages := 0
	if total > 0 {
		totalPages = (total + limit - 1) / limit
	}

	JSON(w, http.StatusOK, PaginatedResponse{
		Data:       monitors,
		Total:      total,
		Page:       page,
		Limit:      limit,
		TotalPages: totalPages,
	})
}

func (h *MonitorHandler) GetMonitor(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		JSONError(w, http.StatusUnauthorized, "unauthorized", "authentication required")
		return
	}

	monitorID, err := uuid.Parse(chi.URLParam(r, "monitorID"))
	if err != nil {
		JSONError(w, http.StatusBadRequest, "bad_request", "invalid monitor ID format")
		return
	}

	monitor, err := h.repo.GetByIDAndUserID(r.Context(), monitorID, userID)
	if err != nil {
		if errors.Is(err, repository.ErrMonitorNotFound) {
			JSONError(w, http.StatusNotFound, "not_found", "monitor not found")
			return
		}
		JSONError(w, http.StatusInternalServerError, "internal_error", "failed to fetch monitor")
		return
	}

	JSON(w, http.StatusOK, monitor)
}

func (h *MonitorHandler) UpdateMonitor(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		JSONError(w, http.StatusUnauthorized, "unauthorized", "authentication required")
		return
	}

	monitorID, err := uuid.Parse(chi.URLParam(r, "monitorID"))
	if err != nil {
		JSONError(w, http.StatusBadRequest, "bad_request", "invalid monitor ID format")
		return
	}

	monitor, err := h.repo.GetByIDAndUserID(r.Context(), monitorID, userID)
	if err != nil {
		if errors.Is(err, repository.ErrMonitorNotFound) {
			JSONError(w, http.StatusNotFound, "not_found", "monitor not found")
			return
		}
		JSONError(w, http.StatusInternalServerError, "internal_error", "failed to fetch monitor")
		return
	}

	var req UpdateMonitorRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		JSONError(w, http.StatusBadRequest, "bad_request", "invalid JSON payload")
		return
	}

	if err := req.Validate(); err != nil {
		JSONError(w, http.StatusUnprocessableEntity, "validation_error", err.Error())
		return
	}

	if req.Name != nil {
		monitor.Name = *req.Name
	}
	if req.URL != nil {
		monitor.URL = *req.URL
	}
	if req.Method != nil {
		monitor.Method = *req.Method
	}
	if req.IntervalSeconds != nil {
		monitor.IntervalSeconds = *req.IntervalSeconds
	}
	if req.TimeoutSeconds != nil {
		monitor.TimeoutSeconds = *req.TimeoutSeconds
	}
	if monitor.TimeoutSeconds > monitor.IntervalSeconds {
		JSONError(w, http.StatusUnprocessableEntity, "validation_error", "timeout_seconds cannot be greater than interval_seconds")
		return
	}
	if req.ExpectedStatusCode != nil {
		monitor.ExpectedStatusCode = *req.ExpectedStatusCode
	}
	if req.Headers != nil {
		monitor.Headers = req.Headers
	}
	if req.Body != nil {
		monitor.Body = *req.Body
	}
	if req.SSLCheckEnabled != nil {
		monitor.SSLCheckEnabled = *req.SSLCheckEnabled
	}
	if req.FailureThreshold != nil {
		monitor.FailureThreshold = *req.FailureThreshold
	}
	if req.RecoveryThreshold != nil {
		monitor.RecoveryThreshold = *req.RecoveryThreshold
	}

	if err := h.repo.Update(r.Context(), monitor); err != nil {
		JSONError(w, http.StatusInternalServerError, "internal_error", "failed to update monitor")
		return
	}

	JSON(w, http.StatusOK, monitor)
}

func (h *MonitorHandler) DeleteMonitor(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		JSONError(w, http.StatusUnauthorized, "unauthorized", "authentication required")
		return
	}

	monitorID, err := uuid.Parse(chi.URLParam(r, "monitorID"))
	if err != nil {
		JSONError(w, http.StatusBadRequest, "bad_request", "invalid monitor ID format")
		return
	}

	if err := h.repo.Delete(r.Context(), monitorID, userID); err != nil {
		if errors.Is(err, repository.ErrMonitorNotFound) {
			JSONError(w, http.StatusNotFound, "not_found", "monitor not found")
			return
		}
		JSONError(w, http.StatusInternalServerError, "internal_error", "failed to delete monitor")
		return
	}

	JSONMessage(w, http.StatusOK, "monitor deleted successfully")
}

func (h *MonitorHandler) EnableMonitor(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		JSONError(w, http.StatusUnauthorized, "unauthorized", "authentication required")
		return
	}

	monitorID, err := uuid.Parse(chi.URLParam(r, "monitorID"))
	if err != nil {
		JSONError(w, http.StatusBadRequest, "bad_request", "invalid monitor ID format")
		return
	}

	if err := h.repo.SetEnabled(r.Context(), monitorID, userID, true); err != nil {
		if errors.Is(err, repository.ErrMonitorNotFound) {
			JSONError(w, http.StatusNotFound, "not_found", "monitor not found")
			return
		}
		JSONError(w, http.StatusInternalServerError, "internal_error", "failed to enable monitor")
		return
	}

	JSONMessage(w, http.StatusOK, "monitor enabled successfully")
}

func (h *MonitorHandler) DisableMonitor(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		JSONError(w, http.StatusUnauthorized, "unauthorized", "authentication required")
		return
	}

	monitorID, err := uuid.Parse(chi.URLParam(r, "monitorID"))
	if err != nil {
		JSONError(w, http.StatusBadRequest, "bad_request", "invalid monitor ID format")
		return
	}

	if err := h.repo.SetEnabled(r.Context(), monitorID, userID, false); err != nil {
		if errors.Is(err, repository.ErrMonitorNotFound) {
			JSONError(w, http.StatusNotFound, "not_found", "monitor not found")
			return
		}
		JSONError(w, http.StatusInternalServerError, "internal_error", "failed to disable monitor")
		return
	}

	JSONMessage(w, http.StatusOK, "monitor disabled successfully")
}

func (h *MonitorHandler) GetResults(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		JSONError(w, http.StatusUnauthorized, "unauthorized", "authentication required")
		return
	}

	monitorID, err := uuid.Parse(chi.URLParam(r, "monitorID"))
	if err != nil {
		JSONError(w, http.StatusBadRequest, "bad_request", "invalid monitor ID format")
		return
	}

	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page < 1 {
		page = 1
	}
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit < 1 || limit > 100 {
		limit = 50
	}

	results, total, err := h.repo.GetResults(r.Context(), monitorID, userID, page, limit)
	if err != nil {
		if errors.Is(err, repository.ErrMonitorNotFound) {
			JSONError(w, http.StatusNotFound, "not_found", "monitor not found")
			return
		}
		JSONError(w, http.StatusInternalServerError, "internal_error", "failed to fetch monitor results")
		return
	}

	if results == nil {
		results = []models.Heartbeat{}
	}

	totalPages := 0
	if total > 0 {
		totalPages = (total + limit - 1) / limit
	}

	JSON(w, http.StatusOK, PaginatedResponse{
		Data:       results,
		Total:      total,
		Page:       page,
		Limit:      limit,
		TotalPages: totalPages,
	})
}

func (h *MonitorHandler) GetIncidents(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		JSONError(w, http.StatusUnauthorized, "unauthorized", "authentication required")
		return
	}

	monitorID, err := uuid.Parse(chi.URLParam(r, "monitorID"))
	if err != nil {
		JSONError(w, http.StatusBadRequest, "bad_request", "invalid monitor ID format")
		return
	}

	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page < 1 {
		page = 1
	}
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit < 1 || limit > 100 {
		limit = 20
	}

	incidents, total, err := h.repo.GetIncidents(r.Context(), monitorID, userID, page, limit)
	if err != nil {
		if errors.Is(err, repository.ErrMonitorNotFound) {
			JSONError(w, http.StatusNotFound, "not_found", "monitor not found")
			return
		}
		JSONError(w, http.StatusInternalServerError, "internal_error", "failed to fetch incidents")
		return
	}

	if incidents == nil {
		incidents = []models.Incident{}
	}

	totalPages := 0
	if total > 0 {
		totalPages = (total + limit - 1) / limit
	}

	JSON(w, http.StatusOK, PaginatedResponse{
		Data:       incidents,
		Total:      total,
		Page:       page,
		Limit:      limit,
		TotalPages: totalPages,
	})
}

func (h *MonitorHandler) GetDashboardSummary(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		JSONError(w, http.StatusUnauthorized, "unauthorized", "authentication required")
		return
	}

	summary, err := h.repo.GetDashboardSummary(r.Context(), userID)
	if err != nil {
		JSONError(w, http.StatusInternalServerError, "internal_error", "failed to calculate dashboard summary")
		return
	}

	JSON(w, http.StatusOK, summary)
}

func (h *MonitorHandler) ListAlertChannels(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		JSONError(w, http.StatusUnauthorized, "unauthorized", "authentication required")
		return
	}

	channels, err := h.repo.ListAlertChannels(r.Context(), userID)
	if err != nil {
		JSONError(w, http.StatusInternalServerError, "internal_error", "failed to fetch alert channels")
		return
	}

	if channels == nil {
		channels = []models.AlertChannel{}
	}

	JSON(w, http.StatusOK, channels)
}

func (h *MonitorHandler) CreateAlertChannel(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		JSONError(w, http.StatusUnauthorized, "unauthorized", "authentication required")
		return
	}

	var req CreateAlertChannelRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		JSONError(w, http.StatusBadRequest, "bad_request", "invalid JSON payload")
		return
	}

	if err := req.Validate(); err != nil {
		JSONError(w, http.StatusUnprocessableEntity, "validation_error", err.Error())
		return
	}

	channel := models.AlertChannel{
		ID:        uuid.New(),
		UserID:    userID,
		Type:      req.Type,
		Name:      req.Name,
		Config:    req.Config,
		IsEnabled: true,
	}

	if err := h.repo.CreateAlertChannel(r.Context(), &channel); err != nil {
		JSONError(w, http.StatusInternalServerError, "internal_error", "failed to create alert channel")
		return
	}

	JSON(w, http.StatusCreated, channel)
}

func (h *MonitorHandler) DeleteAlertChannel(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		JSONError(w, http.StatusUnauthorized, "unauthorized", "authentication required")
		return
	}

	channelID, err := uuid.Parse(chi.URLParam(r, "channelID"))
	if err != nil {
		JSONError(w, http.StatusBadRequest, "bad_request", "invalid alert channel ID format")
		return
	}

	if err := h.repo.DeleteAlertChannel(r.Context(), channelID, userID); err != nil {
		if errors.Is(err, repository.ErrAlertChannelNotFound) {
			JSONError(w, http.StatusNotFound, "not_found", "alert channel not found")
			return
		}
		JSONError(w, http.StatusInternalServerError, "internal_error", "failed to delete alert channel")
		return
	}

	JSONMessage(w, http.StatusOK, "alert channel deleted successfully")
}
