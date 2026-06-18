package handler

import (
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/aniruddh/uptime-monitor/backend/internal/domain"
	"github.com/aniruddh/uptime-monitor/backend/internal/service"
)

// MonitorHandler exposes monitor registration/management over HTTP.
type MonitorHandler struct {
	monitors *service.MonitorService
}

func NewMonitorHandler(monitors *service.MonitorService) *MonitorHandler {
	return &MonitorHandler{monitors: monitors}
}

type createMonitorRequest struct {
	URL         string `json:"url" binding:"required"`
	Name        string `json:"name"`
	IntervalSec int32  `json:"interval_sec"`
}

type monitorResponse struct {
	ID          int64  `json:"id"`
	URL         string `json:"url"`
	Name        string `json:"name"`
	IntervalSec int32  `json:"interval_sec"`
	CreatedAt   string `json:"created_at"`
}

func toMonitorResponse(m domain.Monitor) monitorResponse {
	return monitorResponse{
		ID:          m.ID,
		URL:         m.URL,
		Name:        m.Name,
		IntervalSec: m.IntervalSec,
		CreatedAt:   m.CreatedAt.Format(time.RFC3339),
	}
}

func (h *MonitorHandler) Create(c *gin.Context) {
	var req createMonitorRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	m, err := h.monitors.RegisterMonitor(c.Request.Context(), req.URL, req.Name, req.IntervalSec)
	if err != nil {
		if errors.Is(err, service.ErrInvalidURL) {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to register monitor"})
		return
	}
	c.JSON(http.StatusCreated, toMonitorResponse(m))
}

func (h *MonitorHandler) List(c *gin.Context) {
	monitors, err := h.monitors.ListMonitors(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list monitors"})
		return
	}
	out := make([]monitorResponse, 0, len(monitors))
	for _, m := range monitors {
		out = append(out, toMonitorResponse(m))
	}
	c.JSON(http.StatusOK, out)
}

func (h *MonitorHandler) Delete(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid monitor id"})
		return
	}
	if err := h.monitors.DeleteMonitor(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete monitor"})
		return
	}
	c.Status(http.StatusNoContent)
}
