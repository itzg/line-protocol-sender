/*

Package sender provides a client that sends Influx line protocol metrics to a TCP endpoint.
The client provides options for batching metrics by size and/or timeout.

A simple implementation of protocol.Metric is also provided to ensure this package is ready-to-use
with no Influx specific implementation needed.

Example

The following would send a metric immediately to the telegraf socket_listener input plugin
listening on port 8094:

	client, err := sender.NewClient(context.Background(), sender.Config{Endpoint: "telegraf:8094"})

	metric := sender.NewSimpleMetric("metric_name")
	metric.AddTag("tag", "t1")
	metric.AddField("intField", 1)
	metric.AddField("floatField", 3.14)
	client.Send(metric)

*/
package sender
