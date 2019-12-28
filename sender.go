package sender

import (
	"context"
	"errors"
	"fmt"
	protocol "github.com/influxdata/line-protocol"
	"net"
	"sync"
	"time"
)

const MetricsChanSize = 100

type ErrorListener func(err error)

type Config struct {
	Endpoint     string
	BatchSize    int
	BatchTimeout time.Duration
	ErrorListener
}

type Client interface {
	Send(m protocol.Metric)
	Flush()
}

func NewClient(ctx context.Context, config Config) (Client, error) {
	if config.Endpoint == "" {
		return nil, errors.New("endpoint is required")
	}
	return &clientImpl{
		ctx:    ctx,
		config: config,
	}, nil
}

type clientImpl struct {
	ctx         context.Context
	config      Config
	metrics     chan protocol.Metric
	processSync sync.Once
}

func (c *clientImpl) Flush() {
	c.metrics <- nil
}

func (c *clientImpl) Send(m protocol.Metric) {
	c.processSync.Do(func() {
		c.metrics = make(chan protocol.Metric, MetricsChanSize)
		go c.processMetrics()
	})

	c.metrics <- m
}

func (c *clientImpl) processMetrics() {

	batch := make([]protocol.Metric, 0, c.config.BatchSize)

	var batchTimerChan <-chan time.Time

	for {
		reset := false

		select {
		case <-c.ctx.Done():
			return

		case m := <-c.metrics:
			if m == nil {
				c.flush(batch)
				reset = true
			} else {
				batch = append(batch, m)
				if c.shouldFlush(len(batch)) {
					c.flush(batch)
					reset = true
				} else if batchTimerChan == nil && c.config.BatchTimeout != 0 {
					batchTimerChan = time.After(c.config.BatchTimeout)
				}
			}

		case <-batchTimerChan:
			c.flush(batch)
			reset = true
		}

		if reset {
			batch = batch[0:0]
			// and "clear" timer
			batchTimerChan = nil
		}
	}
}

func (c *clientImpl) shouldFlush(currentBatchSize int) bool {
	if c.config.BatchSize == 0 && c.config.BatchTimeout == 0 {
		return true
	}

	if c.config.BatchSize > 0 && currentBatchSize >= c.config.BatchSize {
		return true
	}
	return false
}

func (c *clientImpl) flush(batch []protocol.Metric) {
	conn, err := net.Dial("tcp", c.config.Endpoint)
	if err != nil {
		c.reportError(fmt.Errorf("failed to connect: %w", err))
		return
	}

	encoder := protocol.NewEncoder(conn)
	for _, metric := range batch {
		_, err := encoder.Encode(metric)
		if err != nil {
			c.reportError(fmt.Errorf("failed to encode: %w", err))
		}
	}

	err = conn.Close()
	if err != nil {
		c.reportError(fmt.Errorf("failed to close: %w", err))
	}
}

func (c *clientImpl) reportError(err error) {
	if c.config.ErrorListener != nil {
		c.config.ErrorListener(err)
	}
}
