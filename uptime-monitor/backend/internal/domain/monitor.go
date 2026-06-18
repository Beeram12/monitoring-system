package domain

import (
	"context"
	"time"
)

// Monitor is a registered URL that gets periodically health-checked.
type Monitor struct {
	ID          int64
	URL         string
	Name        string
	IntervalSec int32
	CreatedAt   time.Time
}

// Check is the result of a single health-check ping against a Monitor.
type Check struct {
	ID         int64
	MonitorID  int64
	StatusCode int32
	ResponseMs int32
	IsUp       bool
	Error      string
	CheckedAt  time.Time
}

// MonitorStatus is a Monitor joined with its most recent Check, used by the API/UI.
type MonitorStatus struct {
	Monitor     Monitor
	LatestCheck *Check
}

// MonitorRepository persists and retrieves monitors.
type MonitorRepository interface {
	Create(ctx context.Context, url, name string, intervalSec int32) (Monitor, error)
	Get(ctx context.Context, id int64) (Monitor, error)
	List(ctx context.Context) ([]Monitor, error)
	Delete(ctx context.Context, id int64) error
}

// CheckRepository persists and retrieves health-check results.
type CheckRepository interface {
	Create(ctx context.Context, c Check) (Check, error)
	ListByMonitor(ctx context.Context, monitorID int64, limit int32) ([]Check, error)
	LatestByMonitor(ctx context.Context, monitorID int64) (*Check, error)
	LatestForAllMonitors(ctx context.Context) ([]Check, error)
}

// Pinger performs a single HTTP health-check against a URL.
type Pinger interface {
	Ping(ctx context.Context, url string) Check
}
