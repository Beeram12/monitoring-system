package routes

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/aniruddh/uptime-monitor/backend/internal/handler"
)

// Register wires every API route to its handler. This is the single place
// that defines the HTTP surface of the application.
func Register(r *gin.Engine, monitorHandler *handler.MonitorHandler, checkHandler *handler.CheckHandler) {
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	api := r.Group("/api")
	{
		api.GET("/monitors", monitorHandler.List)
		api.POST("/monitors", monitorHandler.Create)
		api.DELETE("/monitors/:id", monitorHandler.Delete)

		api.GET("/monitors/:id/checks", checkHandler.ListChecks)
		api.GET("/statuses", checkHandler.ListStatuses)
	}
}
