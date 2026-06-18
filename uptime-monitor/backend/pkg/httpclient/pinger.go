package httpclient

import (
	"context"
	"net/http"
	"time"

	"github.com/aniruddh/uptime-monitor/backend/internal/domain"
)

// Pinger performs HTTP GET health-checks with a bounded timeout.
type Pinger struct {
	client  *http.Client
	timeout time.Duration
}

func NewPinger(timeout time.Duration) *Pinger {
	return &Pinger{
		client: &http.Client{
			// Each request gets its own context-derived deadline; this is a ceiling.
			Timeout: timeout,
		},
		timeout: timeout,
	}
}

var _ domain.Pinger = (*Pinger)(nil)

func (p *Pinger) Ping(ctx context.Context, url string) domain.Check {
	ctx, cancel := context.WithTimeout(ctx, p.timeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return domain.Check{IsUp: false, Error: err.Error()}
	}

	start := time.Now()
	resp, err := p.client.Do(req)
	elapsed := time.Since(start)

	if err != nil {
		return domain.Check{
			IsUp:       false,
			ResponseMs: int32(elapsed.Milliseconds()),
			Error:      err.Error(),
		}
	}
	defer resp.Body.Close()

	return domain.Check{
		StatusCode: int32(resp.StatusCode),
		ResponseMs: int32(elapsed.Milliseconds()),
		IsUp:       resp.StatusCode >= 200 && resp.StatusCode < 400,
	}
}
