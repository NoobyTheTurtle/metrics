package metrics

import (
	"fmt"
	"math/rand"
	"net/http"
	"time"
)

type GaugeMetric string

type CounterMetric string
type Metrics struct {
	Gauges    map[GaugeMetric]float64
	Counters  map[CounterMetric]int64
	serverURL string
	logger    metricsLogger
	random    *rand.Rand
	client    *http.Client
}

func NewMetrics(serverAddress string, log metricsLogger, useTLS bool) *Metrics {
	protocol := "http"

	if useTLS {
		protocol = "https"
	}

	serverURL := fmt.Sprintf("%s://%s", protocol, serverAddress)

	source := rand.NewSource(time.Now().UnixNano())
	random := rand.New(source)

	return &Metrics{
		Gauges:    make(map[GaugeMetric]float64),
		Counters:  make(map[CounterMetric]int64),
		serverURL: serverURL,
		logger:    log,
		random:    random,
		client:    &http.Client{},
	}
}
