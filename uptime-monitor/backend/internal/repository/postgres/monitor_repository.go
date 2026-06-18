package postgres

import (
	"context"

	"github.com/aniruddh/uptime-monitor/backend/internal/db/sqlc"
	"github.com/aniruddh/uptime-monitor/backend/internal/domain"
)

// MonitorRepository is a Postgres-backed implementation of domain.MonitorRepository.
type MonitorRepository struct {
	q *sqlc.Queries
}

func NewMonitorRepository(q *sqlc.Queries) *MonitorRepository {
	return &MonitorRepository{q: q}
}

var _ domain.MonitorRepository = (*MonitorRepository)(nil)

func (r *MonitorRepository) Create(ctx context.Context, url, name string, intervalSec int32) (domain.Monitor, error) {
	m, err := r.q.CreateMonitor(ctx, sqlc.CreateMonitorParams{
		Url:         url,
		Name:        name,
		IntervalSec: intervalSec,
	})
	if err != nil {
		return domain.Monitor{}, err
	}
	return toDomainMonitor(m), nil
}

func (r *MonitorRepository) Get(ctx context.Context, id int64) (domain.Monitor, error) {
	m, err := r.q.GetMonitor(ctx, id)
	if err != nil {
		return domain.Monitor{}, err
	}
	return toDomainMonitor(m), nil
}

func (r *MonitorRepository) List(ctx context.Context) ([]domain.Monitor, error) {
	rows, err := r.q.ListMonitors(ctx)
	if err != nil {
		return nil, err
	}
	out := make([]domain.Monitor, 0, len(rows))
	for _, m := range rows {
		out = append(out, toDomainMonitor(m))
	}
	return out, nil
}

func (r *MonitorRepository) Delete(ctx context.Context, id int64) error {
	return r.q.DeleteMonitor(ctx, id)
}
