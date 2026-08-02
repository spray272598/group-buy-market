package mq

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

// Config RabbitMQ 配置（对齐 Java spring.rabbitmq）
type Config struct {
	URL          string // amqp://admin:admin@127.0.0.1:5672/
	Exchange     string // group_buy_market_exchange
	TeamSuccess  TopicConfig
	TeamRefund   TopicConfig
}

type TopicConfig struct {
	RoutingKey string
	Queue      string
}

// Client RabbitMQ 客户端：声明交换机/队列并提供发布与消费
type Client struct {
	cfg  Config
	conn *amqp.Connection
	ch   *amqp.Channel
}

func NewClient(cfg Config) (*Client, error) {
	if cfg.URL == "" {
		cfg.URL = "amqp://admin:admin@127.0.0.1:5672/"
	}
	if cfg.Exchange == "" {
		cfg.Exchange = "group_buy_market_exchange"
	}
	if cfg.TeamSuccess.RoutingKey == "" {
		cfg.TeamSuccess.RoutingKey = "topic.team_success"
	}
	if cfg.TeamSuccess.Queue == "" {
		cfg.TeamSuccess.Queue = "group_buy_market_queue_2_topic_team_success"
	}
	if cfg.TeamRefund.RoutingKey == "" {
		cfg.TeamRefund.RoutingKey = "topic.team_refund"
	}
	if cfg.TeamRefund.Queue == "" {
		cfg.TeamRefund.Queue = "group_buy_market_queue_2_topic_team_refund"
	}

	conn, err := amqp.Dial(cfg.URL)
	if err != nil {
		return nil, fmt.Errorf("rabbitmq dial: %w", err)
	}
	ch, err := conn.Channel()
	if err != nil {
		_ = conn.Close()
		return nil, fmt.Errorf("rabbitmq channel: %w", err)
	}

	c := &Client{cfg: cfg, conn: conn, ch: ch}
	if err := c.declare(); err != nil {
		_ = c.Close()
		return nil, err
	}
	return c, nil
}

func (c *Client) declare() error {
	if err := c.ch.ExchangeDeclare(c.cfg.Exchange, "topic", true, false, false, false, nil); err != nil {
		return err
	}
	for _, t := range []TopicConfig{c.cfg.TeamSuccess, c.cfg.TeamRefund} {
		q, err := c.ch.QueueDeclare(t.Queue, true, false, false, false, nil)
		if err != nil {
			return err
		}
		if err := c.ch.QueueBind(q.Name, t.RoutingKey, c.cfg.Exchange, false, nil); err != nil {
			return err
		}
	}
	return c.ch.Qos(1, 0, false)
}

// Publish 发布持久化消息到 topic exchange
func (c *Client) Publish(ctx context.Context, routingKey, body string) error {
	return c.ch.PublishWithContext(ctx, c.cfg.Exchange, routingKey, false, false, amqp.Publishing{
		ContentType:  "application/json",
		Body:         []byte(body),
		DeliveryMode: amqp.Persistent,
		Timestamp:    time.Now(),
	})
}

// Consume 启动消费循环（阻塞于 goroutine 中调用）
func (c *Client) Consume(queue string, handler func(body []byte) error) error {
	deliveries, err := c.ch.Consume(queue, "", false, false, false, false, nil)
	if err != nil {
		return err
	}
	go func() {
		for d := range deliveries {
			if err := handler(d.Body); err != nil {
				slog.Error("MQ消费失败，消息Nack重试", "queue", queue, "err", err)
				_ = d.Nack(false, true)
				continue
			}
			_ = d.Ack(false)
		}
	}()
	return nil
}

func (c *Client) TeamSuccessQueue() string { return c.cfg.TeamSuccess.Queue }
func (c *Client) TeamRefundQueue() string  { return c.cfg.TeamRefund.Queue }
func (c *Client) TeamSuccessKey() string   { return c.cfg.TeamSuccess.RoutingKey }
func (c *Client) TeamRefundKey() string    { return c.cfg.TeamRefund.RoutingKey }
func (c *Client) Exchange() string         { return c.cfg.Exchange }

func (c *Client) Close() error {
	if c.ch != nil {
		_ = c.ch.Close()
	}
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}

// EventPublisher 领域出站：发送领域事件
type EventPublisher struct {
	client *Client
}

func NewEventPublisher(client *Client) *EventPublisher {
	return &EventPublisher{client: client}
}

func (p *EventPublisher) Publish(ctx context.Context, routingKey, message string) error {
	if p == nil || p.client == nil {
		slog.Warn("MQ未就绪，消息仅落日志", "routingKey", routingKey, "message", message)
		return nil
	}
	if err := p.client.Publish(ctx, routingKey, message); err != nil {
		slog.Error("发送MQ消息失败", "routingKey", routingKey, "err", err)
		return err
	}
	slog.Info("发送MQ消息成功", "routingKey", routingKey)
	return nil
}
