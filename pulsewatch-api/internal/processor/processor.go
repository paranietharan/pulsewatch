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

	if triggerIncidentOpen || triggerIncidentResolve {
		eventType := "OPEN"
		if triggerIncidentResolve {
			eventType = "RESOLVED"
		}
		if err := queueIncidentAlerts(ctx, tx, m, incidentID, eventType, causeSummary, incidentStartedAt, result.CheckedAt, incidentDuration); err != nil {
			return fmt.Errorf("queue incident alerts: %w", err)
		}
	}
	return tx.Commit(ctx)
}

func queueIncidentAlerts(ctx context.Context, tx pgx.Tx, m models.Monitor, incidentID uuid.UUID, eventType, cause string, startedAt, checkedAt time.Time, duration int) error {
	rows, err := tx.Query(ctx, `SELECT ac.id,ac.type,ac.config FROM alert_channels ac
        JOIN monitor_alert_channels mac ON mac.alert_channel_id=ac.id
        WHERE mac.monitor_id=$1 AND ac.is_enabled`, m.ID)
	if err != nil {
		return err
	}
	channels, err := pgx.CollectRows(rows, func(row pgx.CollectableRow) (models.AlertChannel, error) {
		var ch models.AlertChannel
		err := row.Scan(&ch.ID, &ch.Type, &ch.Config)
		return ch, err
	})
	if err != nil {
		return err
	}
	if len(channels) == 0 {
		var email string
		if err := tx.QueryRow(ctx, "SELECT email FROM users WHERE id=$1", m.UserID).Scan(&email); err != nil {
			return err
		}
		cfg, err := json.Marshal(map[string]string{"email": email})
		if err != nil {
			return err
		}
		channels = append(channels, models.AlertChannel{Type: "EMAIL", Config: cfg})
	}
	for _, ch := range channels {
		var cfg struct{ Email, URL, Secret string }
		if err := json.Unmarshal(ch.Config, &cfg); err != nil {
			return err
		}
		id := uuid.NewSHA1(incidentID, []byte(eventType+"/"+ch.ID.String()))
		var key string
		var payload any
		switch ch.Type {
		case "EMAIL":
			if cfg.Email == "" {
				continue
			}
			if eventType == "OPEN" {
				key = rabbitmq.RoutingKeyEmailIncidentOpened
				payload = rabbitmq.IncidentOpenedEmailPayload{MonitorID: m.ID, MonitorName: m.Name, MonitorURL: m.URL, Email: cfg.Email, Cause: cause, StartedAt: startedAt}
			} else {
				key = rabbitmq.RoutingKeyEmailIncidentResolved
				payload = rabbitmq.IncidentResolvedEmailPayload{MonitorID: m.ID, MonitorName: m.Name, Email: cfg.Email, DurationSeconds: duration, ResolvedAt: checkedAt}
			}
		case "WEBHOOK":
			if cfg.URL == "" {
				continue
			}
			_, err := tx.Exec(ctx, `INSERT INTO alert_deliveries(id,incident_id,alert_channel_id,channel_type,status,attempt_count,updated_at)
                VALUES($1,$2,$3,'WEBHOOK','PENDING',0,now()) ON CONFLICT(id) DO NOTHING`, id, incidentID, ch.ID)
			if err != nil {
				return err
			}
			key = rabbitmq.RoutingKeyWebhookDispatch
			hook := rabbitmq.WebhookDispatchPayload{DeliveryID: id, ChannelID: ch.ID, TargetURL: cfg.URL, Secret: cfg.Secret,
				EventType: "incident." + stringsToLower(eventType), IncidentID: incidentID, MonitorID: m.ID, MonitorName: m.Name,
				Status: eventType, CauseSummary: cause, StartedAt: startedAt, Timestamp: checkedAt}
			if eventType == "RESOLVED" {
				hook.ResolvedAt = &checkedAt
				hook.DurationSeconds = &duration
			}
			payload = hook
		default:
			continue
		}
		data, err := json.Marshal(payload)
		if err != nil {
			return err
		}
		_, err = tx.Exec(ctx, `INSERT INTO auth_outbox_events(id,event_type,routing_key,payload)
            VALUES($1,$2,$3,$4) ON CONFLICT(id) DO NOTHING`, id, eventType, key, data)
		if err != nil {
			return err
		}
	}
	return nil
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
