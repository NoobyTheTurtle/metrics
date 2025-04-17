package model

//go:generate easyjson -all snapshot.go
type MemSnapshot struct {
	Gauges   map[string]float64 `json:"gauges"`
	Counters map[string]int64   `json:"counters"`
}
