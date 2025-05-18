package metric

import (
	"fmt"
	"time"

	"github.com/shirou/gopsutil/v4/cpu"
	"github.com/shirou/gopsutil/v4/mem"
)

func (m *Metrics) InitGomutiMetrics(pollInterval time.Duration) error {
	_, err := cpu.Percent(pollInterval, true)
	return err
}

func (m *Metrics) CollectGopsutilMetrics() error {

	vmStat, err := mem.VirtualMemory()
	if err != nil {
		return fmt.Errorf("metric.CollectGopsutilMetrics: failed to get virtual memory stats: %w", err)
	}
	m.Gauges[GaugeMetric("TotalMemory")] = float64(vmStat.Total)
	m.Gauges[GaugeMetric("FreeMemory")] = float64(vmStat.Free)

	cpuPercentages, err := cpu.Percent(0, true)
	if err != nil {
		return fmt.Errorf("metric.CollectGopsutilMetrics: failed to get cpu percent: %w", err)
	}
	for i, cpuPercent := range cpuPercentages {
		metricName := fmt.Sprintf("CPUutilization%d", i+1)
		m.Gauges[GaugeMetric(metricName)] = cpuPercent
	}

	return nil
}
