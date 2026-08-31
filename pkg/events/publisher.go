package events

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
)

type Event struct {
	ID       string    `json:"id"`
	Type     string    `json:"type"`
	Data     any       `json:"data"`
	Occurred time.Time `json:"occurred"`
}

type Publisher struct {
	nc *nats.Conn
	js jetstream.JetStream
}

func NewPublisher(url string) (*Publisher, error) {
	nc, err := nats.Connect(url)
	if err != nil {
		return nil, fmt.Errorf("connect to nats: %w", err)
	}

	js, err := jetstream.New(nc)
	if err != nil {
		return nil, fmt.Errorf("create jetstream context: %w", err)
	}

	return &Publisher{nc: nc, js: js}, nil
}

func (p *Publisher) Publish(ctx context.Context, subject string, data any) error {
	payload, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("marshal event: %w", err)
	}

	_, err = p.js.Publish(ctx, subject, payload)
	if err != nil {
		return fmt.Errorf("publish to %s: %w", subject, err)
	}
	return nil
}

func (p *Publisher) Close() {
	p.nc.Close()
}

func (p *Publisher) Conn() *nats.Conn {
	return p.nc
}

func (p *Publisher) JetStream() jetstream.JetStream {
	return p.js
}
