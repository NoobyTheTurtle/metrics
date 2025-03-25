package metrics

import (
	"math/rand"
	"time"
)

type GaugeMetric string

const (
	Alloc         GaugeMetric = "Alloc"
	BuckHashSys   GaugeMetric = "BuckHashSys"
	Frees         GaugeMetric = "Frees"
	GCCPUFraction GaugeMetric = "GCCPUFraction"
	GCSys         GaugeMetric = "GCSys"
	HeapAlloc     GaugeMetric = "HeapAlloc"
	HeapIdle      GaugeMetric = "HeapIdle"
	HeapInuse     GaugeMetric = "HeapInuse"
	HeapObjects   GaugeMetric = "HeapObjects"
	HeapReleased  GaugeMetric = "HeapReleased"
	HeapSys       GaugeMetric = "HeapSys"
	LastGC        GaugeMetric = "LastGC"
	Lookups       GaugeMetric = "Lookups"
	MCacheInuse   GaugeMetric = "MCacheInuse"
	MCacheSys     GaugeMetric = "MCacheSys"
	MSpanInuse    GaugeMetric = "MSpanInuse"
	MSpanSys      GaugeMetric = "MSpanSys"
	Mallocs       GaugeMetric = "Mallocs"
	NextGC        GaugeMetric = "NextGC"
	NumForcedGC   GaugeMetric = "NumForcedGC"
	NumGC         GaugeMetric = "NumGC"
	OtherSys      GaugeMetric = "OtherSys"
	PauseTotalNs  GaugeMetric = "PauseTotalNs"
	StackInuse    GaugeMetric = "StackInuse"
	StackSys      GaugeMetric = "StackSys"
	Sys           GaugeMetric = "Sys"
	TotalAlloc    GaugeMetric = "TotalAlloc"
	RandomValue   GaugeMetric = "RandomValue"
)

type CounterMetric string

const (
	PollCount CounterMetric = "PollCount"
)

type Metrics struct {
	Gauges        map[GaugeMetric]float64
	Counters      map[CounterMetric]int64
	random        *rand.Rand
	serverAddress string
}

func NewMetrics(serverAddress string) *Metrics {
	source := rand.NewSource(time.Now().UnixNano())
	random := rand.New(source)

	return &Metrics{
		Gauges:        make(map[GaugeMetric]float64),
		Counters:      make(map[CounterMetric]int64),
		random:        random,
		serverAddress: serverAddress,
	}
}
