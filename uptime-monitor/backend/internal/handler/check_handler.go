package handler

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/aniruddh/uptime-monitor/backend/internal/domain"
	"github.com/aniruddh/uptime-monitor/backend/internal/service"
)

// CheckHandler exposes health-check history and monitor status over HTTP.
type CheckHandler struct {
	checks   *service.CheckService
	monitors *service.MonitorService
}

func NewCheckHandler(checks *service.CheckService, monitors *service.MonitorService) *CheckHandler {
	return &CheckHandler{checks: checks, monitors: monitors}
}

type checkResponse struct {
	ID         int64  `json:"id"`
	MonitorID  int64  `json:"monitor_id"`
	StatusCode int32  `json:"status_code"`
	ResponseMs int32  `json:"response_ms"`
	IsUp       bool   `json:"is_up"`
	Error      string `json:"error"`
	CheckedAt  string `json:"checked_at"`
}

type monitorStatusResponse struct {
	Monitor     monitorResponse `json:"monitor"`
	LatestCheck *checkResponse  `json:"latest_check"`
}

func toCheckResponse(c domain.Check) checkResponse {
	return checkResponse{
		ID:         c.ID,
		MonitorID:  c.MonitorID,
		StatusCode: c.StatusCode,
		ResponseMs: c.ResponseMs,
		IsUp:       c.IsUp,
		Error:      c.Error,
		CheckedAt:  c.CheckedAt.Format(time.RFC3339),
	}
}

func (h *CheckHandler) ListChecks(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid monitor id"})
		return
	}
	checks, err := h.checks.ListChecks(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list checks"})
		return
	}
	out := make([]checkResponse, 0, len(checks))
	for _, ch := range checks {
		out = append(out, toCheckResponse(ch))
	}
	c.JSON(http.StatusOK, out)
}

func (h *CheckHandler) ListStatuses(c *gin.Context) {
	statuses, err := h.monitors.ListStatuses(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list statuses"})
		return
	}
	out := make([]monitorStatusResponse, 0, len(statuses))
	for _, s := range statuses {
		resp := monitorStatusResponse{Monitor: toMonitorResponse(s.Monitor)}
		if s.LatestCheck != nil {
			cr := toCheckResponse(*s.LatestCheck)
			resp.LatestCheck = &cr
		}
		out = append(out, resp)
	}
	c.JSON(http.StatusOK, out)
}
