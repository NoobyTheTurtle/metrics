package metric

import (
	"math/rand"
	"time"
)

type GaugeMetric string

type CounterMetric string

type Metrics struct {
	Gauges    map[GaugeMetric]float64
	Counters  map[CounterMetric]int64
	logger    MetricsLogger
	random    *rand.Rand
	transport MetricsTransport
}

func NewMetrics(serverAddress string, log MetricsLogger, useTLS bool, key string, encrypter Encrypter) *Metrics {
	httpTransport := NewHTTPTransport(serverAddress, useTLS, key, encrypter, log)

	return NewMetricsWithTransport(log, httpTransport)
}

func NewMetricsWithTransport(log MetricsLogger, transport MetricsTransport) *Metrics {
	source := rand.NewSource(time.Now().UnixNano())
	random := rand.New(source)

	return &Metrics{
		Gauges:    make(map[GaugeMetric]float64),
		Counters:  make(map[CounterMetric]int64),
		logger:    log,
		random:    random,
		transport: transport,
	}
}
