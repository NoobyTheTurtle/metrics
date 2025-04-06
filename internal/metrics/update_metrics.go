package metrics

import (
	"runtime"
)

func (m *Metrics) UpdateMetrics() {
	m.updateGaugeMemStats()
	m.updateGaugeRandomValue()
	m.updateCounters()
}

func (m *Metrics) updateGaugeMemStats() {
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)

	for _, metric := range MemStatsMetrics {
		m.Gauges[metric.Metric] = metric.GetValue(&memStats)
	}
}

func (m *Metrics) updateGaugeRandomValue() {
	m.Gauges[RandomValue] = m.random.Float64()
}

func (m *Metrics) updateCounters() {
	m.Counters[PollCount]++
}
