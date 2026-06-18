package service

import (
	"context"
	"errors"

	"github.com/aniruddh/uptime-monitor/backend/internal/domain"
)

var ErrInvalidURL = errors.New("url must be a non-empty http(s) URL")

const (
	defaultIntervalSec = 60
	minIntervalSec     = 5
)

// MonitorService implements the application's business logic around
// registering monitors and reading their status, decoupled from HTTP and storage.
type MonitorService struct {
	monitors domain.MonitorRepository
	checks   domain.CheckRepository
}

func NewMonitorService(monitors domain.MonitorRepository, checks domain.CheckRepository) *MonitorService {
	return &MonitorService{monitors: monitors, checks: checks}
}

func (s *MonitorService) RegisterMonitor(ctx context.Context, url, name string, intervalSec int32) (domain.Monitor, error) {
	if url == "" {
		return domain.Monitor{}, ErrInvalidURL
	}
	if intervalSec <= 0 {
		intervalSec = defaultIntervalSec
	}
	if intervalSec < minIntervalSec {
		intervalSec = minIntervalSec
	}
	return s.monitors.Create(ctx, url, name, intervalSec)
}

func (s *MonitorService) ListMonitors(ctx context.Context) ([]domain.Monitor, error) {
	return s.monitors.List(ctx)
}

func (s *MonitorService) DeleteMonitor(ctx context.Context, id int64) error {
	return s.monitors.Delete(ctx, id)
}

// ListStatuses returns every monitor paired with its most recent health check.
func (s *MonitorService) ListStatuses(ctx context.Context) ([]domain.MonitorStatus, error) {
	monitors, err := s.monitors.List(ctx)
	if err != nil {
		return nil, err
	}

	latest, err := s.checks.LatestForAllMonitors(ctx)
	if err != nil {
		return nil, err
	}
	byMonitor := make(map[int64]domain.Check, len(latest))
	for _, c := range latest {
		byMonitor[c.MonitorID] = c
	}

	out := make([]domain.MonitorStatus, 0, len(monitors))
	for _, m := range monitors {
		status := domain.MonitorStatus{Monitor: m}
		if c, ok := byMonitor[m.ID]; ok {
			status.LatestCheck = &c
		}
		out = append(out, status)
	}
	return out, nil
}
