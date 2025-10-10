package model

import (
	"math/rand/v2"
	"runtime"
	"sync/atomic"
	"time"
)

type Collection struct {
	counter atomic.Int64
}

func NewCollection() *Collection {
	return &Collection{}
}

func (c *Collection) Collect() []Metric {
	c.counter.Add(1)

	rnd := rand.New(rand.NewPCG(uint64(time.Now().UnixNano()), 0))

	var ms runtime.MemStats
	runtime.ReadMemStats(&ms)

	return []Metric{
		Gauge("Alloc", float64(ms.Alloc)),
		Gauge("BuckHashSys", float64(ms.BuckHashSys)),
		Gauge("Frees", float64(ms.Frees)),
		Gauge("GCCPUFraction", ms.GCCPUFraction),
		Gauge("GCSys", float64(ms.GCSys)),
		Gauge("HeapAlloc", float64(ms.HeapAlloc)),
		Gauge("HeapIdle", float64(ms.HeapIdle)),
		Gauge("HeapInuse", float64(ms.HeapInuse)),
		Gauge("HeapObjects", float64(ms.HeapObjects)),
		Gauge("HeapReleased", float64(ms.HeapReleased)),
		Gauge("HeapSys", float64(ms.HeapSys)),
		Gauge("LastGC", float64(ms.LastGC)),
		Gauge("Lookups", float64(ms.Lookups)),
		Gauge("MCacheInuse", float64(ms.MCacheInuse)),
		Gauge("MCacheSys", float64(ms.MCacheSys)),
		Gauge("MSpanInuse", float64(ms.MSpanInuse)),
		Gauge("MSpanSys", float64(ms.MSpanSys)),
		Gauge("Mallocs", float64(ms.Mallocs)),
		Gauge("NextGC", float64(ms.NextGC)),
		Gauge("NumForcedGC", float64(ms.NumForcedGC)),
		Gauge("NumGC", float64(ms.NumGC)),
		Gauge("OtherSys", float64(ms.OtherSys)),
		Gauge("PauseTotalNs", float64(ms.PauseTotalNs)),
		Gauge("StackInuse", float64(ms.StackInuse)),
		Gauge("StackSys", float64(ms.StackSys)),
		Gauge("Sys", float64(ms.Sys)),
		Gauge("TotalAlloc", float64(ms.TotalAlloc)),
		Gauge("RandomValue", rnd.Float64()),
		Counter("PollCount", c.counter.Load()),
	}
}
