package processor

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"pulsewatch-api/internal/models"
	"pulsewatch-api/internal/rabbitmq"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	amqp "github.com/rabbitmq/amqp091-go"
)

type ResultProcessor struct {
	db        *pgxpool.Pool
	rmqClient *rabbitmq.Client
	logger    *slog.Logger
}

func NewResultProcessor(db *pgxpool.Pool, rmqClient *rabbitmq.Client, logger *slog.Logger) *ResultProcessor {
	return &ResultProcessor{
		db:        db,
		rmqClient: rmqClient,
		logger:    logger,
	}
}

// Start launches the consumer loops for monitor results and delivery outcomes.
func (p *ResultProcessor) Start(ctx context.Context) error {
	if p.rmqClient == nil {
		return fmt.Errorf("rabbitmq client is nil")
	}
	conn := p.rmqClient.Connection()
	if conn == nil {
		return fmt.Errorf("rabbitmq connection is nil")
	}

	resultsCh, err := conn.Channel()
	if err != nil {
		return fmt.Errorf("failed to open results channel: %w", err)
	}

	if err := resultsCh.Qos(20, 0, false); err != nil {
		resultsCh.Close()
		return fmt.Errorf("failed to set results QoS: %w", err)
	}

	resultMsgs, err := resultsCh.Consume(
		rabbitmq.ResultsQueue,
		"pulsewatch-api-result-processor",
		false,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		resultsCh.Close()
		return fmt.Errorf("failed to consume from results queue: %w", err)
	}

	// Delivery outcomes consumer channel
	outcomesCh, err := conn.Channel()
	if err != nil {
		resultsCh.Close()
		return fmt.Errorf("failed to open outcomes channel: %w", err)
	}

	outcomeMsgs, err := outcomesCh.Consume(
		rabbitmq.DeliveryOutcomeQueue,
		"pulsewatch-api-outcome-processor",
		false,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		resultsCh.Close()
		outcomesCh.Close()
		return fmt.Errorf("failed to consume from outcomes queue: %w", err)
	}

	p.logger.Info("Result and outcome processors listening on RabbitMQ")

	go func() {
		<-ctx.Done()
		_ = resultsCh.Close()
		_ = outcomesCh.Close()
	}()

	// Handle delivery outcomes in a goroutine
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case d, ok := <-outcomeMsgs:
				if !ok {
					return
				}
				p.processOutcomeDelivery(ctx, d)
			}
		}
	}()

	// Handle check results in main loop
	for {
		select {
		case <-ctx.Done():
			return nil
		case d, ok := <-resultMsgs:
			if !ok {
				return nil
			}
			p.processResultDelivery(ctx, d)
		}
	}
}

func (p *ResultProcessor) processOutcomeDelivery(ctx context.Context, d amqp.Delivery) {
	var payload rabbitmq.DeliveryOutcomePayload
	if err := json.Unmarshal(d.Body, &payload); err != nil {
		p.logger.Error("Failed to deserialize delivery outcome payload", "error", err)
		_ = d.Nack(false, false)
		return
	}

	query := `
		UPDATE alert_deliveries
		SET status = $1,
		    attempt_count = $2,
		    last_error = $3,
		    updated_at = NOW()
		WHERE id = $4;
	`
	_, err := p.db.Exec(ctx, query, payload.Status, payload.AttemptCount, payload.LastError, payload.DeliveryID)
	if err != nil {
		p.logger.Error("Failed to update alert delivery in database", "delivery_id", payload.DeliveryID, "error", err)
		_ = d.Nack(false, true)
		return
	}

	_ = d.Ack(false)
}

func (p *ResultProcessor) processResultDelivery(ctx context.Context, d amqp.Delivery) {
	var result rabbitmq.CheckResult
	if err := json.Unmarshal(d.Body, &result); err != nil {
		p.logger.Error("Failed to deserialize check result body, dead-lettering", "error", err)
		_ = d.Nack(false, false)
		return
	}

	if err := p.ProcessResult(ctx, result); err != nil {
		p.logger.Error("Failed to process check result, requeuing message",
			"monitor_id", result.MonitorID,
			"job_id", result.JobID,
			"error", err,
		)
		_ = d.Nack(false, true)
		return
	}

	_ = d.Ack(false)
}

// ProcessResult processes a single check result idempotently.
func (p *ResultProcessor) ProcessResult(ctx context.Context, result rabbitmq.CheckResult) error {
	// Idempotency Check / Deduplication
	if result.JobID != uuid.Nil {
		var exists bool
		err := p.db.QueryRow(ctx, "SELECT EXISTS(SELECT 1 FROM heartbeats WHERE job_id = $1)", result.JobID).Scan(&exists)
		if err == nil && exists {
			p.logger.Debug("Duplicate check result detected, ignoring", "job_id", result.JobID)
			return nil
		}
	}

	return p.processResultTx(ctx, result)
}

func (p *ResultProcessor) processResultTx(ctx context.Context, result rabbitmq.CheckResult) error {
	tx, err := p.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	// Insert Heartbeat record
	insertHeartbeatQuery := `
		INSERT INTO heartbeats (
			monitor_id, status, status_code, response_time_ms, dns_lookup_ms,
			tcp_connect_ms, tls_handshake_ms, error_message, region, checked_at, job_id
		) VALUES ($1, $2::heartbeat_status, $3, $4, $5, $6, $7, $8, $9, $10, $11);
	`
	_, err = tx.Exec(ctx, insertHeartbeatQuery,
		result.MonitorID,
		result.Status,
		result.StatusCode,
		result.ResponseTimeMs,
		result.DNSLookupMs,
		result.TCPConnectMs,
		result.TLSHandshakeMs,
		result.ErrorMessage,
		result.Region,
		result.CheckedAt,
		result.JobID,
	)
	if err != nil {
		return fmt.Errorf("failed to insert heartbeat: %w", err)
	}

	// Fetch and lock monitor state
	selectMonitorQuery := `
		SELECT id, user_id, name, url, method, status, is_active,
		       failure_threshold, recovery_threshold, consecutive_failures,
		       consecutive_successes, avg_response_time_ms
		FROM monitors
		WHERE id = $1
		FOR UPDATE;
	`
	var m models.Monitor
	err = tx.QueryRow(ctx, selectMonitorQuery, result.MonitorID).Scan(
		&m.ID,
		&m.UserID,
		&m.Name,
		&m.URL,
		&m.Method,
		&m.Status,
		&m.IsActive,
		&m.FailureThreshold,
		&m.RecoveryThreshold,
		&m.ConsecutiveFailures,
		&m.ConsecutiveSuccesses,
		&m.AvgResponseTimeMs,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return tx.Commit(ctx)
		}
		return fmt.Errorf("failed to select monitor: %w", err)
	}

	// Update consecutive counters
	if result.Status == "UP" {
		m.ConsecutiveSuccesses++
		m.ConsecutiveFailures = 0
	} else {
		m.ConsecutiveFailures++
		m.ConsecutiveSuccesses = 0
	}

	// Rolling average calculation
	newAvg := result.ResponseTimeMs
	if m.AvgResponseTimeMs > 0 && result.ResponseTimeMs > 0 {
		newAvg = (m.AvgResponseTimeMs*9 + result.ResponseTimeMs) / 10
	}

	newStatus := m.Status
	var triggerIncidentOpen bool
	var triggerIncidentResolve bool
	var incidentID uuid.UUID
	var incidentStartedAt time.Time
	var incidentDuration int
	var causeSummary string

	// Incident State Machine Evaluation
	if m.ConsecutiveFailures >= m.FailureThreshold {
		var activeID uuid.UUID
		err := tx.QueryRow(ctx, `
			SELECT id FROM incidents
			WHERE monitor_id = $1 AND status IN ('OPEN', 'ACKNOWLEDGED')
			LIMIT 1;
		`, m.ID).Scan(&activeID)

		if err == pgx.ErrNoRows {
			incidentID = uuid.New()
			causeSummary = result.ErrorMessage
			if causeSummary == "" {
				causeSummary = fmt.Sprintf("Health check failed with status %d", result.StatusCode)
			}
			incidentStartedAt = result.CheckedAt

			insertIncidentQuery := `
				INSERT INTO incidents (id, monitor_id, status, cause_summary, started_at)
				VALUES ($1, $2, 'OPEN'::incident_status, $3, $4);
			`
			if _, err := tx.Exec(ctx, insertIncidentQuery, incidentID, m.ID, causeSummary, incidentStartedAt); err != nil {
				return fmt.Errorf("failed to insert incident: %w", err)
			}

			insertUpdateQuery := `
				INSERT INTO incident_updates (incident_id, status, message)
				VALUES ($1, 'OPEN'::incident_update_status, $2);
			`
			if _, err := tx.Exec(ctx, insertUpdateQuery, incidentID, causeSummary); err != nil {
				return fmt.Errorf("failed to insert incident update: %w", err)
			}

			newStatus = "DOWN"
			triggerIncidentOpen = true
		}
	} else if m.ConsecutiveSuccesses >= m.RecoveryThreshold && m.Status == "DOWN" {
		err := tx.QueryRow(ctx, `
			SELECT id, started_at FROM incidents
			WHERE monitor_id = $1 AND status IN ('OPEN', 'ACKNOWLEDGED')
			ORDER BY started_at DESC
			LIMIT 1;
		`, m.ID).Scan(&incidentID, &incidentStartedAt)

		if err == nil {
			duration := int(result.CheckedAt.Sub(incidentStartedAt).Seconds())
			incidentDuration = duration

			resolveQuery := `
				UPDATE incidents
				SET status = 'RESOLVED'::incident_status,
				    resolved_at = $1,
				    duration_seconds = $2,
				    updated_at = NOW()
				WHERE id = $3;
			`
			if _, err := tx.Exec(ctx, resolveQuery, result.CheckedAt, duration, incidentID); err != nil {
				return fmt.Errorf("failed to resolve incident: %w", err)
			}

			insertUpdateQuery := `
				INSERT INTO incident_updates (incident_id, status, message)
				VALUES ($1, 'RESOLVED'::incident_update_status, 'All health checks passing, threshold met');
			`
			if _, err := tx.Exec(ctx, insertUpdateQuery, incidentID); err != nil {
				return fmt.Errorf("failed to insert incident resolve update: %w", err)
			}

			newStatus = "UP"
			triggerIncidentResolve = true
		} else {
			newStatus = "UP"
		}
	} else if m.ConsecutiveSuccesses > 0 && m.Status == "PENDING" {
		newStatus = "UP"
	}

	// Update Monitor Row
	updateMonitorQuery := `
		UPDATE monitors
		SET status = $1::monitor_status,
		    last_checked_at = $2,
		    consecutive_failures = $3,
		    consecutive_successes = $4,
		    avg_response_time_ms = $5,
		    updated_at = NOW()
		WHERE id = $6;
	`
	_, err = tx.Exec(ctx, updateMonitorQuery,
		newStatus,
		result.CheckedAt,
		m.ConsecutiveFailures,
		m.ConsecutiveSuccesses,
		newAvg,
		m.ID,
	)
	if err != nil {
		return fmt.Errorf("failed to update monitor: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	// Asynchronous Alert Dispatches over RabbitMQ
	if triggerIncidentOpen {
		p.logger.Warn("Triggering incident opened alert",
			"monitor_name", m.Name,
			"incident_id", incidentID,
			"cause", causeSummary,
		)
		go p.dispatchAlerts(context.Background(), m, incidentID, "OPEN", causeSummary, incidentStartedAt, nil, nil)
	}

	if triggerIncidentResolve {
		p.logger.Info("Triggering incident resolved alert",
			"monitor_name", m.Name,
			"incident_id", incidentID,
			"duration_sec", incidentDuration,
		)
		resolvedAt := result.CheckedAt
		go p.dispatchAlerts(context.Background(), m, incidentID, "RESOLVED", "", incidentStartedAt, &resolvedAt, &incidentDuration)
	}

	return nil
}

func (p *ResultProcessor) dispatchAlerts(
	ctx context.Context,
	m models.Monitor,
	incidentID uuid.UUID,
	eventType string, // OPEN, RESOLVED
	cause string,
	startedAt time.Time,
	resolvedAt *time.Time,
	durationSeconds *int,
) {
	// Query user notification destinations and associated alert channels
	rows, err := p.db.Query(ctx, `
		SELECT ac.id, ac.type, ac.name, ac.config, ac.is_enabled
		FROM alert_channels ac
		JOIN monitor_alert_channels mac ON mac.alert_channel_id = ac.id
		WHERE mac.monitor_id = $1 AND ac.is_enabled = TRUE;
	`, m.ID)
	if err != nil {
		p.logger.Error("Failed to query alert channels for monitor", "monitor_id", m.ID, "error", err)
		return
	}
	defer rows.Close()

	var channels []models.AlertChannel
	for rows.Next() {
		var ac models.AlertChannel
		if err := rows.Scan(&ac.ID, &ac.Type, &ac.Name, &ac.Config, &ac.IsEnabled); err == nil {
			channels = append(channels, ac)
		}
	}

	// Fallback to owner email if no specific channels bound
	var ownerEmail string
	_ = p.db.QueryRow(ctx, "SELECT email FROM users WHERE id = $1", m.UserID).Scan(&ownerEmail)

	if len(channels) == 0 && ownerEmail != "" {
		p.publishEmailAlert(ctx, ownerEmail, m, eventType, cause, startedAt, resolvedAt, durationSeconds)
		return
	}

	for _, ch := range channels {
		switch ch.Type {
		case "EMAIL":
			var cfg struct {
				Email string `json:"email"`
			}
			if err := json.Unmarshal(ch.Config, &cfg); err == nil && cfg.Email != "" {
				p.publishEmailAlert(ctx, cfg.Email, m, eventType, cause, startedAt, resolvedAt, durationSeconds)
			}
		case "WEBHOOK":
			var cfg struct {
				URL    string `json:"url"`
				Secret string `json:"secret,omitempty"`
			}
			if err := json.Unmarshal(ch.Config, &cfg); err == nil && cfg.URL != "" {
				p.publishWebhookAlert(ctx, ch.ID, cfg.URL, cfg.Secret, m, incidentID, eventType, cause, startedAt, resolvedAt, durationSeconds)
			}
		}
	}
}

func (p *ResultProcessor) publishEmailAlert(
	ctx context.Context,
	email string,
	m models.Monitor,
	eventType string,
	cause string,
	startedAt time.Time,
	resolvedAt *time.Time,
	durationSeconds *int,
) {
	if p.rmqClient == nil {
		return
	}
	if eventType == "OPEN" {
		payload := rabbitmq.IncidentOpenedEmailPayload{
			MonitorID:   m.ID,
			MonitorName: m.Name,
			MonitorURL:  m.URL,
			Email:       email,
			Cause:       cause,
			StartedAt:   startedAt,
		}
		data, err := json.Marshal(payload)
		if err != nil {
			p.logger.Error("Failed to serialize incident opened email payload", "error", err)
			return
		}
		_ = p.rmqClient.PublishConfirm(ctx, rabbitmq.RoutingKeyEmailIncidentOpened, data)
	} else if eventType == "RESOLVED" && resolvedAt != nil && durationSeconds != nil {
		payload := rabbitmq.IncidentResolvedEmailPayload{
			MonitorID:       m.ID,
			MonitorName:     m.Name,
			MonitorURL:      m.URL,
			Email:           email,
			DurationSeconds: *durationSeconds,
			ResolvedAt:      *resolvedAt,
		}
		data, err := json.Marshal(payload)
		if err != nil {
			p.logger.Error("Failed to serialize incident resolved email payload", "error", err)
			return
		}
		_ = p.rmqClient.PublishConfirm(ctx, rabbitmq.RoutingKeyEmailIncidentResolved, data)
	}
}

func (p *ResultProcessor) publishWebhookAlert(
	ctx context.Context,
	channelID uuid.UUID,
	targetURL, secret string,
	m models.Monitor,
	incidentID uuid.UUID,
	eventType string,
	cause string,
	startedAt time.Time,
	resolvedAt *time.Time,
	durationSeconds *int,
) {
	deliveryID := uuid.New()

	// Record initial PENDING delivery state in alert_deliveries
	_, _ = p.db.Exec(ctx, `
		INSERT INTO alert_deliveries (id, incident_id, alert_channel_id, channel_type, status, attempt_count, updated_at)
		VALUES ($1, $2, $3, 'WEBHOOK', 'PENDING', 0, NOW())
		ON CONFLICT DO NOTHING;
	`, deliveryID, incidentID, channelID)

	if p.rmqClient == nil {
		return
	}

	payload := rabbitmq.WebhookDispatchPayload{
		DeliveryID:      deliveryID,
		ChannelID:       channelID,
		TargetURL:       targetURL,
		Secret:          secret,
		EventType:       fmt.Sprintf("incident.%s", stringsToLower(eventType)),
		IncidentID:      incidentID,
		MonitorID:       m.ID,
		MonitorName:     m.Name,
		Status:          eventType,
		CauseSummary:    cause,
		StartedAt:       startedAt,
		ResolvedAt:      resolvedAt,
		DurationSeconds: durationSeconds,
		Timestamp:       time.Now().UTC(),
	}

	data, err := json.Marshal(payload)
	if err != nil {
		p.logger.Error("Failed to serialize webhook dispatch payload", "error", err)
		return
	}

	_ = p.rmqClient.PublishConfirm(ctx, rabbitmq.RoutingKeyWebhookDispatch, data)
}

func stringsToLower(s string) string {
	switch s {
	case "OPEN":
		return "opened"
	case "RESOLVED":
		return "resolved"
	default:
		return s
	}
}
