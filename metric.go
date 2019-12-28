package sender

import (
	protocol "github.com/influxdata/line-protocol"
	"time"
)

type SimpleMetric struct {
	name      string
	tags      []*protocol.Tag
	fields    []*protocol.Field
	timestamp time.Time
}

func NewSimpleMetric(name string) *SimpleMetric {
	return &SimpleMetric{name: name}
}

func (m *SimpleMetric) SetTime(t time.Time) {
	m.timestamp = t
}

func (m *SimpleMetric) Time() time.Time {
	if m.timestamp.IsZero() {
		return time.Now()
	} else {
		return m.timestamp
	}
}

func (m *SimpleMetric) Name() string {
	return m.name
}

func (m *SimpleMetric) TagList() []*protocol.Tag {
	return m.tags
}

func (m *SimpleMetric) FieldList() []*protocol.Field {
	return m.fields
}

func (m *SimpleMetric) AddTag(key, value string) {
	m.tags = append(m.tags, &protocol.Tag{
		Key:   key,
		Value: value,
	})
}

func (m *SimpleMetric) AddField(key string, value interface{}) {
	m.fields = append(m.fields, &protocol.Field{
		Key:   key,
		Value: value,
	})
}
