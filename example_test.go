package sender_test

import (
	"bytes"
	"context"
	"fmt"
	sender "github.com/itzg/line-protocol-sender"
	"io"
	"log"
	"net"
	"time"
)

type ExampleEndpoint struct {
	listener net.Listener
}

func NewExampleEndpoint() *ExampleEndpoint {
	listener, err := net.Listen("tcp", "127.0.0.1:")
	if err != nil {
		log.Fatal(err)
	}
	e := &ExampleEndpoint{listener: listener}
	go e.listen()
	return e
}

func (e *ExampleEndpoint) Addr() string {
	return e.listener.Addr().String()
}

func (e *ExampleEndpoint) listen() {
	conn, err := e.listener.Accept()
	if err != nil {
		log.Fatal(err)
	}
	var buffer bytes.Buffer
	_, err = io.Copy(&buffer, conn)
	if err != nil {
		log.Fatal(err)
	}
	conn.Close()

	fmt.Print(buffer.String())
}

func Example_sending() {
	endpoint := NewExampleEndpoint()

	client, _ := sender.NewClient(context.Background(), sender.Config{Endpoint: endpoint.Addr()})

	metric := sender.NewSimpleMetric("metric_name")
	metric.SetTime(time.Unix(3, 1))
	metric.AddTag("tag", "t1")
	metric.AddField("intField", 1)
	metric.AddField("floatField", 3.14)
	client.Send(metric)

	// allow time for listener to receive line
	time.Sleep(10 * time.Millisecond)

	//Output:
	//metric_name,tag=t1 intField=1i,floatField=3.14 3000000001
}
