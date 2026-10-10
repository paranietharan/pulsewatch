package repository

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"pulsewatch-api/internal/models"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	ErrMonitorNotFound      = errors.New("monitor not found")
	ErrAlertChannelNotFound = errors.New("alert channel not found")
)

type MonitorRepository struct {
	db *pgxpool.Pool
}

func NewMonitorRepository(db *pgxpool.Pool) *MonitorRepository {
	return &MonitorRepository{db: db}
}

func (r *MonitorRepository) Create(ctx context.Context, m *models.Monitor, channelIDs []uuid.UUID) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	headersJSON, err := json.Marshal(m.Headers)
	if err != nil {
		headersJSON = []byte("{}")
	}

	query := `
		INSERT INTO monitors (
			id, user_id, name, url, method, interval_seconds, timeout_seconds,
			expected_status_code, headers, body, failure_threshold, recovery_threshold,
			status, is_active, ssl_check_enabled, next_check_at, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5::http_method, $6, $7,
			$8, $9, $10, $11, $12,
			$13::monitor_status, $14, $15, $16, $17, $18
		) RETURNING created_at, updated_at;
	`

	now := time.Now().UTC()
	m.CreatedAt = now
	m.UpdatedAt = now
	if m.NextCheckAt.IsZero() {
		m.NextCheckAt = now
	}

	err = tx.QueryRow(ctx, query,
		m.ID,
		m.UserID,
		m.Name,
		m.URL,
		m.Method,
		m.IntervalSeconds,
		m.TimeoutSeconds,
		m.ExpectedStatusCode,
		headersJSON,
		m.Body,
		m.FailureThreshold,
		m.RecoveryThreshold,
		m.Status,
		m.IsActive,
		m.SSLCheckEnabled,
		m.NextCheckAt,
		m.CreatedAt,
		m.UpdatedAt,
	).Scan(&m.CreatedAt, &m.UpdatedAt)
	if err != nil {
		return fmt.Errorf("failed to insert monitor: %w", err)
	}

	for _, chID := range channelIDs {
		_, err := tx.Exec(ctx, `
			INSERT INTO monitor_alert_channels (monitor_id, alert_channel_id)
			VALUES ($1, $2)
			ON CONFLICT DO NOTHING;
		`, m.ID, chID)
		if err != nil {
			return fmt.Errorf("failed to bind alert channel to monitor: %w", err)
		}
	}

	return tx.Commit(ctx)
}

func (r *MonitorRepository) ListByUserID(ctx context.Context, userID uuid.UUID, page, limit int, status string) ([]models.Monitor, int, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}
	offset := (page - 1) * limit

	var countQuery string
	var query string
	var args []interface{}
	var countArgs []interface{}

	if status != "" {
		countQuery = "SELECT COUNT(*) FROM monitors WHERE user_id = $1 AND status = $2::monitor_status;"
		countArgs = []interface{}{userID, status}
		query = `
			SELECT id, user_id, name, url, method, interval_seconds, timeout_seconds,
			       expected_status_code, headers, body, status, is_active, ssl_check_enabled,
			       failure_threshold, recovery_threshold, consecutive_failures, consecutive_successes,
			       last_checked_at, next_check_at, uptime_24h, uptime_7d, uptime_30d,
			       avg_response_time_ms, created_at, updated_at
			FROM monitors
			WHERE user_id = $1 AND status = $2::monitor_status
			ORDER BY created_at DESC
			LIMIT $3 OFFSET $4;
		`
		args = []interface{}{userID, status, limit, offset}
	} else {
		countQuery = "SELECT COUNT(*) FROM monitors WHERE user_id = $1;"
		countArgs = []interface{}{userID}
		query = `
			SELECT id, user_id, name, url, method, interval_seconds, timeout_seconds,
			       expected_status_code, headers, body, status, is_active, ssl_check_enabled,
			       failure_threshold, recovery_threshold, consecutive_failures, consecutive_successes,
			       last_checked_at, next_check_at, uptime_24h, uptime_7d, uptime_30d,
			       avg_response_time_ms, created_at, updated_at
			FROM monitors
			WHERE user_id = $1
			ORDER BY created_at DESC
			LIMIT $2 OFFSET $3;
		`
		args = []interface{}{userID, limit, offset}
	}

	var total int
	if err := r.db.QueryRow(ctx, countQuery, countArgs...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("failed to count monitors: %w", err)
	}

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to query monitors: %w", err)
	}
	defer rows.Close()

	var monitors []models.Monitor
	for rows.Next() {
		var m models.Monitor
		var headersRaw []byte
		var body *string

		err := rows.Scan(
			&m.ID,
			&m.UserID,
			&m.Name,
			&m.URL,
			&m.Method,
			&m.IntervalSeconds,
			&m.TimeoutSeconds,
			&m.ExpectedStatusCode,
			&headersRaw,
			&body,
			&m.Status,
			&m.IsActive,
			&m.SSLCheckEnabled,
			&m.FailureThreshold,
			&m.RecoveryThreshold,
			&m.ConsecutiveFailures,
			&m.ConsecutiveSuccesses,
			&m.LastCheckedAt,
			&m.NextCheckAt,
			&m.Uptime24h,
			&m.Uptime7d,
			&m.Uptime30d,
			&m.AvgResponseTimeMs,
			&m.CreatedAt,
			&m.UpdatedAt,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan monitor row: %w", err)
		}

		if len(headersRaw) > 0 {
			_ = json.Unmarshal(headersRaw, &m.Headers)
		}
		if m.Headers == nil {
			m.Headers = make(map[string]string)
		}
		if body != nil {
			m.Body = *body
		}

		monitors = append(monitors, m)
	}

	return monitors, total, rows.Err()
}

func (r *MonitorRepository) GetByIDAndUserID(ctx context.Context, monitorID, userID uuid.UUID) (*models.Monitor, error) {
	query := `
		SELECT id, user_id, name, url, method, interval_seconds, timeout_seconds,
		       expected_status_code, headers, body, status, is_active, ssl_check_enabled,
		       failure_threshold, recovery_threshold, consecutive_failures, consecutive_successes,
		       last_checked_at, next_check_at, uptime_24h, uptime_7d, uptime_30d,
		       avg_response_time_ms, created_at, updated_at
		FROM monitors
		WHERE id = $1 AND user_id = $2;
	`
	var m models.Monitor
	var headersRaw []byte
	var body *string

	err := r.db.QueryRow(ctx, query, monitorID, userID).Scan(
		&m.ID,
		&m.UserID,
		&m.Name,
		&m.URL,
		&m.Method,
		&m.IntervalSeconds,
		&m.TimeoutSeconds,
		&m.ExpectedStatusCode,
		&headersRaw,
		&body,
		&m.Status,
		&m.IsActive,
		&m.SSLCheckEnabled,
		&m.FailureThreshold,
		&m.RecoveryThreshold,
		&m.ConsecutiveFailures,
		&m.ConsecutiveSuccesses,
		&m.LastCheckedAt,
		&m.NextCheckAt,
		&m.Uptime24h,
		&m.Uptime7d,
		&m.Uptime30d,
		&m.AvgResponseTimeMs,
		&m.CreatedAt,
		&m.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrMonitorNotFound
		}
		return nil, fmt.Errorf("failed to query monitor: %w", err)
	}

	if len(headersRaw) > 0 {
		_ = json.Unmarshal(headersRaw, &m.Headers)
	}
	if m.Headers == nil {
		m.Headers = make(map[string]string)
	}
	if body != nil {
		m.Body = *body
	}

	return &m, nil
}

func (r *MonitorRepository) Update(ctx context.Context, m *models.Monitor) error {
	headersJSON, err := json.Marshal(m.Headers)
	if err != nil {
		headersJSON = []byte("{}")
	}

	query := `
		UPDATE monitors
		SET name = $1,
		    url = $2,
		    method = $3::http_method,
		    interval_seconds = $4,
		    timeout_seconds = $5,
		    expected_status_code = $6,
		    headers = $7,
		    body = $8,
		    failure_threshold = $9,
		    recovery_threshold = $10,
		    ssl_check_enabled = $11,
		    updated_at = NOW()
		WHERE id = $12 AND user_id = $13;
	`
	res, err := r.db.Exec(ctx, query,
		m.Name,
		m.URL,
		m.Method,
		m.IntervalSeconds,
		m.TimeoutSeconds,
		m.ExpectedStatusCode,
		headersJSON,
		m.Body,
		m.FailureThreshold,
		m.RecoveryThreshold,
		m.SSLCheckEnabled,
		m.ID,
		m.UserID,
	)
	if err != nil {
		return fmt.Errorf("failed to update monitor: %w", err)
	}
	if res.RowsAffected() == 0 {
		return ErrMonitorNotFound
	}
	return nil
}

func (r *MonitorRepository) SetEnabled(ctx context.Context, monitorID, userID uuid.UUID, enabled bool) error {
	query := `
		UPDATE monitors
		SET is_active = $1,
		    status = CASE WHEN $1 = FALSE THEN 'PENDING'::monitor_status ELSE status END,
		    next_check_at = CASE WHEN $1 = TRUE THEN NOW() ELSE next_check_at END,
		    updated_at = NOW()
		WHERE id = $2 AND user_id = $3;
	`
	res, err := r.db.Exec(ctx, query, enabled, monitorID, userID)
	if err != nil {
		return fmt.Errorf("failed to update monitor status: %w", err)
	}
	if res.RowsAffected() == 0 {
		return ErrMonitorNotFound
	}
	return nil
}

func (r *MonitorRepository) Delete(ctx context.Context, monitorID, userID uuid.UUID) error {
	res, err := r.db.Exec(ctx, "DELETE FROM monitors WHERE id = $1 AND user_id = $2;", monitorID, userID)
	if err != nil {
		return fmt.Errorf("failed to delete monitor: %w", err)
	}
	if res.RowsAffected() == 0 {
		return ErrMonitorNotFound
	}
	return nil
}

func (r *MonitorRepository) GetResults(ctx context.Context, monitorID, userID uuid.UUID, page, limit int) ([]models.Heartbeat, int, error) {
	// First verify monitor belongs to user
	var exists bool
	err := r.db.QueryRow(ctx, "SELECT EXISTS(SELECT 1 FROM monitors WHERE id = $1 AND user_id = $2);", monitorID, userID).Scan(&exists)
	if err != nil || !exists {
		return nil, 0, ErrMonitorNotFound
	}

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 50
	}
	offset := (page - 1) * limit

	var total int
	if err := r.db.QueryRow(ctx, "SELECT COUNT(*) FROM heartbeats WHERE monitor_id = $1;", monitorID).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("failed to count heartbeats: %w", err)
	}

	query := `
		SELECT id, monitor_id, status, status_code, response_time_ms,
		       dns_lookup_ms, tcp_connect_ms, tls_handshake_ms,
		       error_message, region, checked_at, job_id
		FROM heartbeats
		WHERE monitor_id = $1
		ORDER BY checked_at DESC
		LIMIT $2 OFFSET $3;
	`
	rows, err := r.db.Query(ctx, query, monitorID, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to query heartbeats: %w", err)
	}
	defer rows.Close()

	var heartbeats []models.Heartbeat
	for rows.Next() {
		var h models.Heartbeat
		var errMsg *string
		var statusCode *int
		err := rows.Scan(
			&h.ID,
			&h.MonitorID,
			&h.Status,
			&statusCode,
			&h.ResponseTimeMs,
			&h.DNSLookupMs,
			&h.TCPConnectMs,
			&h.TLSHandshakeMs,
			&errMsg,
			&h.Region,
			&h.CheckedAt,
			&h.JobID,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan heartbeat: %w", err)
		}
		if errMsg != nil {
			h.ErrorMessage = *errMsg
		}
		if statusCode != nil {
			h.StatusCode = *statusCode
		}
		heartbeats = append(heartbeats, h)
	}

	return heartbeats, total, rows.Err()
}

func (r *MonitorRepository) GetIncidents(ctx context.Context, monitorID, userID uuid.UUID, page, limit int) ([]models.Incident, int, error) {
	var exists bool
	err := r.db.QueryRow(ctx, "SELECT EXISTS(SELECT 1 FROM monitors WHERE id = $1 AND user_id = $2);", monitorID, userID).Scan(&exists)
	if err != nil || !exists {
		return nil, 0, ErrMonitorNotFound
	}

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}
	offset := (page - 1) * limit

	var total int
	if err := r.db.QueryRow(ctx, "SELECT COUNT(*) FROM incidents WHERE monitor_id = $1;", monitorID).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("failed to count incidents: %w", err)
	}

	query := `
		SELECT i.id, i.monitor_id, m.name, i.status, i.cause_summary,
		       i.started_at, i.resolved_at, i.duration_seconds, i.created_at, i.updated_at
		FROM incidents i
		JOIN monitors m ON m.id = i.monitor_id
		WHERE i.monitor_id = $1
		ORDER BY i.started_at DESC
		LIMIT $2 OFFSET $3;
	`
	rows, err := r.db.Query(ctx, query, monitorID, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to query incidents: %w", err)
	}
	defer rows.Close()

	var incidents []models.Incident
	for rows.Next() {
		var inc models.Incident
		var cause *string
		err := rows.Scan(
			&inc.ID,
			&inc.MonitorID,
			&inc.MonitorName,
			&inc.Status,
			&cause,
			&inc.StartedAt,
			&inc.ResolvedAt,
			&inc.DurationSeconds,
			&inc.CreatedAt,
			&inc.UpdatedAt,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan incident: %w", err)
		}
		if cause != nil {
			inc.CauseSummary = *cause
		}
		incidents = append(incidents, inc)
	}

	return incidents, total, rows.Err()
}

func (r *MonitorRepository) GetDashboardSummary(ctx context.Context, userID uuid.UUID) (*models.DashboardSummary, error) {
	query := `
		SELECT
			COUNT(*) AS total,
			COUNT(*) FILTER (WHERE status = 'UP') AS healthy,
			COUNT(*) FILTER (WHERE status = 'DEGRADED') AS degraded,
			COUNT(*) FILTER (WHERE status = 'DOWN') AS down,
			COUNT(*) FILTER (WHERE is_active = FALSE) AS paused,
			COALESCE(AVG(uptime_24h), 100.0) AS avg_uptime,
			COALESCE(AVG(avg_response_time_ms)::INTEGER, 0) AS avg_latency
		FROM monitors
		WHERE user_id = $1;
	`

	var s models.DashboardSummary
	err := r.db.QueryRow(ctx, query, userID).Scan(
		&s.TotalMonitors,
		&s.HealthyMonitors,
		&s.DegradedMonitors,
		&s.DownMonitors,
		&s.PausedMonitors,
		&s.AvgUptime24h,
		&s.AvgResponseTimeMs,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate dashboard summary: %w", err)
	}

	// Count active incidents across user's monitors
	incidentQuery := `
		SELECT COUNT(*)
		FROM incidents i
		JOIN monitors m ON m.id = i.monitor_id
		WHERE m.user_id = $1 AND i.status IN ('OPEN', 'ACKNOWLEDGED');
	`
	_ = r.db.QueryRow(ctx, incidentQuery, userID).Scan(&s.ActiveIncidents)

	return &s, nil
}

func (r *MonitorRepository) ListAlertChannels(ctx context.Context, userID uuid.UUID) ([]models.AlertChannel, error) {
	query := `
		SELECT id, user_id, type, name, config, is_enabled, created_at, updated_at
		FROM alert_channels
		WHERE user_id = $1
		ORDER BY created_at DESC;
	`
	rows, err := r.db.Query(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to query alert channels: %w", err)
	}
	defer rows.Close()

	var channels []models.AlertChannel
	for rows.Next() {
		var ch models.AlertChannel
		if err := rows.Scan(&ch.ID, &ch.UserID, &ch.Type, &ch.Name, &ch.Config, &ch.IsEnabled, &ch.CreatedAt, &ch.UpdatedAt); err == nil {
			channels = append(channels, ch)
		}
	}
	return channels, rows.Err()
}

func (r *MonitorRepository) CreateAlertChannel(ctx context.Context, ch *models.AlertChannel) error {
	query := `
		INSERT INTO alert_channels (id, user_id, type, name, config, is_enabled)
		VALUES ($1, $2, $3::alert_channel_type, $4, $5, $6)
		RETURNING created_at, updated_at;
	`
	return r.db.QueryRow(ctx, query, ch.ID, ch.UserID, ch.Type, ch.Name, ch.Config, ch.IsEnabled).Scan(&ch.CreatedAt, &ch.UpdatedAt)
}

func (r *MonitorRepository) DeleteAlertChannel(ctx context.Context, channelID, userID uuid.UUID) error {
	res, err := r.db.Exec(ctx, "DELETE FROM alert_channels WHERE id = $1 AND user_id = $2;", channelID, userID)
	if err != nil {
		return fmt.Errorf("failed to delete alert channel: %w", err)
	}
	if res.RowsAffected() == 0 {
		return ErrAlertChannelNotFound
	}
	return nil
}
