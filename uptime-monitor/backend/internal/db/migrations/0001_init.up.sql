CREATE TABLE monitors (
    id              BIGSERIAL PRIMARY KEY,
    url             TEXT NOT NULL,
    name            TEXT NOT NULL DEFAULT '',
    interval_sec    INTEGER NOT NULL DEFAULT 60,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE checks (
    id              BIGSERIAL PRIMARY KEY,
    monitor_id      BIGINT NOT NULL REFERENCES monitors(id) ON DELETE CASCADE,
    status_code     INTEGER NOT NULL DEFAULT 0,
    response_ms     INTEGER NOT NULL DEFAULT 0,
    is_up           BOOLEAN NOT NULL,
    error           TEXT NOT NULL DEFAULT '',
    checked_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_checks_monitor_id_checked_at ON checks (monitor_id, checked_at DESC);
