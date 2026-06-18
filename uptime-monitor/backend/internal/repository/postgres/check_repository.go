package postgres

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"

	"github.com/aniruddh/uptime-monitor/backend/internal/db/sqlc"
	"github.com/aniruddh/uptime-monitor/backend/internal/domain"
)

// CheckRepository is a Postgres-backed implementation of domain.CheckRepository.
type CheckRepository struct {
	q *sqlc.Queries
}

func NewCheckRepository(q *sqlc.Queries) *CheckRepository {
	return &CheckRepository{q: q}
}

var _ domain.CheckRepository = (*CheckRepository)(nil)

func (r *CheckRepository) Create(ctx context.Context, c domain.Check) (domain.Check, error) {
	created, err := r.q.CreateCheck(ctx, sqlc.CreateCheckParams{
		MonitorID:  c.MonitorID,
		StatusCode: c.StatusCode,
		ResponseMs: c.ResponseMs,
		IsUp:       c.IsUp,
		Error:      c.Error,
	})
	if err != nil {
		return domain.Check{}, err
	}
	return toDomainCheck(created), nil
}

func (r *CheckRepository) ListByMonitor(ctx context.Context, monitorID int64, limit int32) ([]domain.Check, error) {
	rows, err := r.q.ListChecksByMonitor(ctx, sqlc.ListChecksByMonitorParams{
		MonitorID: monitorID,
		Limit:     limit,
	})
	if err != nil {
		return nil, err
	}
	out := make([]domain.Check, 0, len(rows))
	for _, c := range rows {
		out = append(out, toDomainCheck(c))
	}
	return out, nil
}

func (r *CheckRepository) LatestByMonitor(ctx context.Context, monitorID int64) (*domain.Check, error) {
	c, err := r.q.GetLatestCheckByMonitor(ctx, monitorID)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	dc := toDomainCheck(c)
	return &dc, nil
}

func (r *CheckRepository) LatestForAllMonitors(ctx context.Context) ([]domain.Check, error) {
	rows, err := r.q.GetLatestChecksForAllMonitors(ctx)
	if err != nil {
		return nil, err
	}
	out := make([]domain.Check, 0, len(rows))
	for _, c := range rows {
		out = append(out, toDomainCheck(c))
	}
	return out, nil
}
