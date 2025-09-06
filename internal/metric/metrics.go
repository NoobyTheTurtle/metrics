package metric

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
	logger    MetricsLogger
	random    *rand.Rand
	client    *http.Client
	key       string
	encrypter Encrypter
}

func NewMetrics(serverAddress string, log MetricsLogger, useTLS bool, key string, encrypter Encrypter) *Metrics {
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
		key:       key,
		encrypter: encrypter,
	}
}
