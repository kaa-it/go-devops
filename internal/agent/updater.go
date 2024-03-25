package agent

import (
	"runtime"
)

func Update(s *Storage) {
	stats := &runtime.MemStats{}

	runtime.ReadMemStats(stats)

	s.UpdateGauge("Alloc", float64(stats.Alloc))
	s.UpdateGauge("BuckHashSys", float64(stats.BuckHashSys))
	s.UpdateGauge("Frees", float64(stats.Frees))
	s.UpdateGauge("GCCPUFraction", float64(stats.GCCPUFraction))
	s.UpdateGauge("GCSys", float64(stats.GCSys))
	s.UpdateGauge("HeapAlloc", float64(stats.HeapAlloc))
	s.UpdateGauge("HeapIdle", float64(stats.HeapIdle))
	s.UpdateGauge("HeapInuse", float64(stats.HeapInuse))
	s.UpdateGauge("HeapObjects", float64(stats.HeapObjects))
	s.UpdateGauge("HeapReleased", float64(stats.HeapReleased))
	s.UpdateGauge("HeapSys", float64(stats.HeapSys))
	s.UpdateGauge("LastGC", float64(stats.LastGC))
	s.UpdateGauge("Lookups", float64(stats.Lookups))
	s.UpdateGauge("MCacheSys", float64(stats.MCacheSys))
	s.UpdateGauge("MCacheInuse", float64(stats.MCacheInuse))
	s.UpdateGauge("MSpanInuse", float64(stats.MSpanInuse))
	s.UpdateGauge("MSpanSys", float64(stats.MSpanSys))
	s.UpdateGauge("Mallocs", float64(stats.Mallocs))
	s.UpdateGauge("NextGC", float64(stats.NextGC))
	s.UpdateGauge("NumForcedGC", float64(stats.NumForcedGC))
	s.UpdateGauge("NumGC", float64(stats.NumGC))
	s.UpdateGauge("OtherSys", float64(stats.OtherSys))
	s.UpdateGauge("PauseTotalNs", float64(stats.PauseTotalNs))
	s.UpdateGauge("StackInuse", float64(stats.StackInuse))
	s.UpdateGauge("StackSys", float64(stats.StackSys))
	s.UpdateGauge("Sys", float64(stats.Sys))
	s.UpdateGauge("TotalAlloc", float64(stats.TotalAlloc))
}
