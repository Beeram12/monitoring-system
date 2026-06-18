package scheduler

import (
	"context"
	"log"
	"sync"
	"time"

	"github.com/aniruddh/uptime-monitor/backend/internal/domain"
)

// Scheduler periodically re-reads the monitor list and runs one ping
// goroutine per monitor at its configured interval. It is safe to call
// Run from a single goroutine; monitors are (re)started automatically when
// new ones are registered.
type Scheduler struct {
	monitors domain.MonitorRepository
	pinger   domain.Pinger
	recorder func(ctx context.Context, monitorID int64, result domain.Check) (domain.Check, error)

	refreshInterval time.Duration

	mu      sync.Mutex
	running map[int64]context.CancelFunc
}

func New(
	monitors domain.MonitorRepository,
	pinger domain.Pinger,
	recorder func(ctx context.Context, monitorID int64, result domain.Check) (domain.Check, error),
	refreshInterval time.Duration,
) *Scheduler {
	return &Scheduler{
		monitors:        monitors,
		pinger:          pinger,
		recorder:        recorder,
		refreshInterval: refreshInterval,
		running:         make(map[int64]context.CancelFunc),
	}
}

// Run blocks until ctx is cancelled, periodically reconciling the set of
// running per-monitor ping loops against what's registered in the database.
func (s *Scheduler) Run(ctx context.Context) {
	s.reconcile(ctx)

	ticker := time.NewTicker(s.refreshInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			s.stopAll()
			return
		case <-ticker.C:
			s.reconcile(ctx)
		}
	}
}

func (s *Scheduler) reconcile(ctx context.Context) {
	monitors, err := s.monitors.List(ctx)
	if err != nil {
		log.Printf("scheduler: failed to list monitors: %v", err)
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	seen := make(map[int64]struct{}, len(monitors))
	for _, m := range monitors {
		seen[m.ID] = struct{}{}
		if _, ok := s.running[m.ID]; ok {
			continue
		}
		mCtx, cancel := context.WithCancel(ctx)
		s.running[m.ID] = cancel
		go s.runLoop(mCtx, m)
	}

	for id, cancel := range s.running {
		if _, ok := seen[id]; !ok {
			cancel()
			delete(s.running, id)
		}
	}
}

func (s *Scheduler) runLoop(ctx context.Context, m domain.Monitor) {
	interval := time.Duration(m.IntervalSec) * time.Second
	if interval <= 0 {
		interval = time.Minute
	}

	s.checkOnce(ctx, m)

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			s.checkOnce(ctx, m)
		}
	}
}

func (s *Scheduler) checkOnce(ctx context.Context, m domain.Monitor) {
	result := s.pinger.Ping(ctx, m.URL)
	if _, err := s.recorder(ctx, m.ID, result); err != nil {
		log.Printf("scheduler: failed to record check for monitor %d: %v", m.ID, err)
	}
}

func (s *Scheduler) stopAll() {
	s.mu.Lock()
	defer s.mu.Unlock()
	for id, cancel := range s.running {
		cancel()
		delete(s.running, id)
	}
}
