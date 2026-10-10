package rabbitmq

import (
	"context"
	"fmt"
	"sync"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

// Client manages connection, channel, and confirmed publishing to RabbitMQ.
type Client struct {
	url      string
	conn     *amqp.Connection
	ch       *amqp.Channel
	confirms chan amqp.Confirmation
	mu       sync.Mutex
	isClosed bool
}

// Connect establishes a connection to RabbitMQ, initializes topology, and enables publisher confirms.
func Connect(url string) (*Client, error) {
	conn, err := amqp.Dial(url)
	if err != nil {
		return nil, fmt.Errorf("failed to dial RabbitMQ: %w", err)
	}

	ch, err := conn.Channel()
	if err != nil {
		_ = conn.Close()
		return nil, fmt.Errorf("failed to open RabbitMQ channel: %w", err)
	}

	// Enable Publisher Confirms
	if err := ch.Confirm(false); err != nil {
		_ = ch.Close()
		_ = conn.Close()
		return nil, fmt.Errorf("failed to enable publisher confirms: %w", err)
	}

	confirms := ch.NotifyPublish(make(chan amqp.Confirmation, 100))

	client := &Client{
		url:      url,
		conn:     conn,
		ch:       ch,
		confirms: confirms,
	}

	// Declare exchange and queues
	if err := client.setupTopology(); err != nil {
		_ = client.Close()
		return nil, fmt.Errorf("failed to setup RabbitMQ topology: %w", err)
	}

	return client, nil
}

// setupTopology declares durable exchange, queues, and routing bindings.
func (c *Client) setupTopology() error {
	// Declare durable direct exchange
	err := c.ch.ExchangeDeclare(
		ExchangeName,
		"direct",
		true,  // durable
		false, // auto-deleted
		false, // internal
		false, // no-wait
		nil,   // arguments
	)
	if err != nil {
		return fmt.Errorf("failed to declare exchange '%s': %w", ExchangeName, err)
	}

	// Declare and bind durable email queue
	if _, err := c.ch.QueueDeclare(EmailQueue, true, false, false, false, nil); err != nil {
		return fmt.Errorf("failed to declare queue '%s': %w", EmailQueue, err)
	}
	emailKeys := []string{
		RoutingKeyEmailVerification,
		RoutingKeyEmailPasswordReset,
		RoutingKeyEmailResendVerification,
		RoutingKeyEmailIncidentOpened,
		RoutingKeyEmailIncidentResolved,
	}
	for _, key := range emailKeys {
		if err := c.ch.QueueBind(EmailQueue, key, ExchangeName, false, nil); err != nil {
			return fmt.Errorf("failed to bind queue '%s' to key '%s': %w", EmailQueue, key, err)
		}
	}

	// Declare and bind durable webhook queue
	if _, err := c.ch.QueueDeclare(WebhookQueue, true, false, false, false, nil); err != nil {
		return fmt.Errorf("failed to declare queue '%s': %w", WebhookQueue, err)
	}
	if err := c.ch.QueueBind(WebhookQueue, RoutingKeyWebhookDispatch, ExchangeName, false, nil); err != nil {
		return fmt.Errorf("failed to bind queue '%s': %w", WebhookQueue, err)
	}

	// Declare and bind durable delivery outcome queue
	if _, err := c.ch.QueueDeclare(DeliveryOutcomeQueue, true, false, false, false, nil); err != nil {
		return fmt.Errorf("failed to declare queue '%s': %w", DeliveryOutcomeQueue, err)
	}
	if err := c.ch.QueueBind(DeliveryOutcomeQueue, RoutingKeyDeliveryOutcome, ExchangeName, false, nil); err != nil {
		return fmt.Errorf("failed to bind queue '%s': %w", DeliveryOutcomeQueue, err)
	}

	// Declare and bind durable monitor results queue
	if _, err := c.ch.QueueDeclare(ResultsQueue, true, false, false, false, nil); err != nil {
		return fmt.Errorf("failed to declare queue '%s': %w", ResultsQueue, err)
	}
	if err := c.ch.QueueBind(ResultsQueue, RoutingKeyResult, ExchangeName, false, nil); err != nil {
		return fmt.Errorf("failed to bind queue '%s': %w", ResultsQueue, err)
	}

	return nil
}

// Connection returns the underlying amqp.Connection.
func (c *Client) Connection() *amqp.Connection {
	return c.conn
}

// PublishConfirm publishes a message and synchronously blocks until broker confirmation is received.
func (c *Client) PublishConfirm(ctx context.Context, routingKey string, body []byte) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.isClosed {
		return fmt.Errorf("rabbitmq client is closed")
	}

	publishing := amqp.Publishing{
		DeliveryMode: amqp.Persistent, // Persistent storage
		ContentType:  "application/json",
		Timestamp:    time.Now().UTC(),
		Body:         body,
	}

	err := c.ch.PublishWithContext(
		ctx,
		ExchangeName,
		routingKey,
		false, // mandatory
		false, // immediate
		publishing,
	)
	if err != nil {
		return fmt.Errorf("failed to publish message to RabbitMQ: %w", err)
	}

	// Wait for publisher confirm
	select {
	case confirmed := <-c.confirms:
		if !confirmed.Ack {
			return fmt.Errorf("message was nacked by rabbitmq broker")
		}
		return nil
	case <-ctx.Done():
		return fmt.Errorf("context cancelled while waiting for publish confirmation: %w", ctx.Err())
	case <-time.After(5 * time.Second):
		return fmt.Errorf("timed out waiting for publish confirmation")
	}
}

// Close gracefully closes the channel and connection.
func (c *Client) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.isClosed = true
	var err error
	if c.ch != nil {
		err = c.ch.Close()
	}
	if c.conn != nil {
		if connErr := c.conn.Close(); connErr != nil && err == nil {
			err = connErr
		}
	}
	return err
}
