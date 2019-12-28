package sender

import (
	"bytes"
	"context"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"io"
	"net"
	"testing"
	"time"
)

type MockEndpoint struct {
	listener net.Listener
	contents []string
	err      error
}

func NewMockEndpoint() (*MockEndpoint, error) {
	listener, err := net.Listen("tcp", "127.0.0.1:")
	if err != nil {
		return nil, err
	}
	e := &MockEndpoint{listener: listener}
	go e.listen()
	return e, nil
}

func (e *MockEndpoint) Addr() string {
	return e.listener.Addr().String()
}

func (e *MockEndpoint) Close() {
	e.listener.Close()
}

func (e *MockEndpoint) listen() {
	for {
		conn, err := e.listener.Accept()
		if err != nil {
			return
		}

		var buffer bytes.Buffer
		_, err = io.Copy(&buffer, conn)
		if err != nil {
			e.err = err
		} else {
			e.contents = append(e.contents, buffer.String())
		}
		e.err = err
		conn.Close()
	}
}

func (e *MockEndpoint) HasContent() bool {
	return len(e.contents) > 0
}

func (e *MockEndpoint) Content() []string {
	return e.contents
}

func TestSendImmediate(t *testing.T) {
	endpoint, err := NewMockEndpoint()
	require.NoError(t, err)
	defer endpoint.Close()

	ctx, cancelFunc := context.WithCancel(context.Background())
	defer cancelFunc()
	client, err := NewClient(ctx, Config{Endpoint: endpoint.Addr()})
	require.NoError(t, err)

	metric := &SimpleMetric{name: "metric_name"}
	metric.SetTime(time.Unix(1, 0))
	metric.AddTag("tag1", "t1")
	metric.AddField("value1", 1)
	client.Send(metric)

	assert.Eventually(t, endpoint.HasContent, 10*time.Millisecond, 1*time.Millisecond)

	assert.Equal(t, "metric_name,tag1=t1 value1=1i 1000000000\n", endpoint.Content()[0])
}

func TestSendImmediate_ResetEachBatch(t *testing.T) {
	endpoint, err := NewMockEndpoint()
	require.NoError(t, err)
	defer endpoint.Close()

	ctx, cancelFunc := context.WithCancel(context.Background())
	defer cancelFunc()
	client, err := NewClient(ctx, Config{Endpoint: endpoint.Addr()})
	require.NoError(t, err)

	metric := &SimpleMetric{name: "metric_name"}
	metric.SetTime(time.Unix(1, 0))
	metric.AddTag("tag1", "t1")
	metric.AddField("value1", 1)
	client.Send(metric)

	metric2 := &SimpleMetric{name: "metric_name"}
	metric2.SetTime(time.Unix(2, 0))
	metric2.AddTag("tag1", "t2")
	metric2.AddField("value1", 2)
	client.Send(metric2)

	assert.Eventually(t, func() bool {
		return len(endpoint.Content()) >= 2
	}, 10*time.Millisecond, 1*time.Millisecond)

	assert.Equal(t, "metric_name,tag1=t1 value1=1i 1000000000\n", endpoint.Content()[0])
	assert.Equal(t, "metric_name,tag1=t2 value1=2i 2000000000\n", endpoint.Content()[1])
}

func TestSendBuffered(t *testing.T) {
	endpoint, err := NewMockEndpoint()
	require.NoError(t, err)
	defer endpoint.Close()

	ctx, cancelFunc := context.WithCancel(context.Background())
	defer cancelFunc()
	client, err := NewClient(ctx, Config{
		Endpoint:  endpoint.Addr(),
		BatchSize: 2,
	})
	require.NoError(t, err)

	metric := &SimpleMetric{name: "metric_name"}
	metric.SetTime(time.Unix(1, 0))
	metric.AddTag("tag1", "t1")
	metric.AddField("value1", 1)
	client.Send(metric)

	time.Sleep(10 * time.Millisecond)
	assert.False(t, endpoint.HasContent())

	metric2 := &SimpleMetric{name: "metric_name"}
	metric2.SetTime(time.Unix(2, 0))
	metric2.AddTag("tag1", "t2")
	metric2.AddField("value1", 2)
	client.Send(metric2)

	assert.Eventually(t, endpoint.HasContent, 10*time.Millisecond, 1*time.Millisecond)

	assert.Equal(t, "metric_name,tag1=t1 value1=1i 1000000000\nmetric_name,tag1=t2 value1=2i 2000000000\n", endpoint.Content()[0])
}

func TestSendBufferedWithFlush(t *testing.T) {
	endpoint, err := NewMockEndpoint()
	require.NoError(t, err)
	defer endpoint.Close()

	ctx, cancelFunc := context.WithCancel(context.Background())
	defer cancelFunc()
	client, err := NewClient(ctx, Config{
		Endpoint:  endpoint.Addr(),
		BatchSize: 2,
	})
	require.NoError(t, err)

	metric := &SimpleMetric{name: "metric_name"}
	metric.SetTime(time.Unix(1, 0))
	metric.AddTag("tag1", "t1")
	metric.AddField("value1", 1)
	client.Send(metric)

	time.Sleep(10 * time.Millisecond)
	assert.False(t, endpoint.HasContent())

	client.Flush()

	assert.Eventually(t, endpoint.HasContent, 10*time.Millisecond, 1*time.Millisecond)

	assert.Equal(t, "metric_name,tag1=t1 value1=1i 1000000000\n", endpoint.Content()[0])
}

func TestSendTimeout(t *testing.T) {
	endpoint, err := NewMockEndpoint()
	require.NoError(t, err)
	defer endpoint.Close()

	ctx, cancelFunc := context.WithCancel(context.Background())
	defer cancelFunc()
	client, err := NewClient(ctx, Config{
		Endpoint:     endpoint.Addr(),
		BatchTimeout: 20 * time.Millisecond,
	})
	require.NoError(t, err)

	metric := &SimpleMetric{name: "metric_name"}
	metric.SetTime(time.Unix(1, 0))
	metric.AddTag("tag1", "t1")
	metric.AddField("value1", 1)
	client.Send(metric)

	time.Sleep(10 * time.Millisecond)
	assert.False(t, endpoint.HasContent())

	assert.Eventually(t, endpoint.HasContent, 30*time.Millisecond, 5*time.Millisecond)

	assert.Equal(t, "metric_name,tag1=t1 value1=1i 1000000000\n", endpoint.Content()[0])
}
