package postgres

import (
	"github.com/aniruddh/uptime-monitor/backend/internal/db/sqlc"
	"github.com/aniruddh/uptime-monitor/backend/internal/domain"
)

func toDomainMonitor(m sqlc.Monitor) domain.Monitor {
	return domain.Monitor{
		ID:          m.ID,
		URL:         m.Url,
		Name:        m.Name,
		IntervalSec: m.IntervalSec,
		CreatedAt:   m.CreatedAt.Time,
	}
}

func toDomainCheck(c sqlc.Check) domain.Check {
	return domain.Check{
		ID:         c.ID,
		MonitorID:  c.MonitorID,
		StatusCode: c.StatusCode,
		ResponseMs: c.ResponseMs,
		IsUp:       c.IsUp,
		Error:      c.Error,
		CheckedAt:  c.CheckedAt.Time,
	}
}
