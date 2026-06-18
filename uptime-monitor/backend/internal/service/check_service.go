package service

import (
	"context"

	"github.com/aniruddh/uptime-monitor/backend/internal/domain"
)

const checksPerMonitorLimit = 50

// CheckService records and retrieves health-check results for monitors.
type CheckService struct {
	checks domain.CheckRepository
}

func NewCheckService(checks domain.CheckRepository) *CheckService {
	return &CheckService{checks: checks}
}

func (s *CheckService) RecordCheck(ctx context.Context, monitorID int64, result domain.Check) (domain.Check, error) {
	result.MonitorID = monitorID
	return s.checks.Create(ctx, result)
}

func (s *CheckService) ListChecks(ctx context.Context, monitorID int64) ([]domain.Check, error) {
	return s.checks.ListByMonitor(ctx, monitorID, checksPerMonitorLimit)
}
