package json2metrics

import (
	"fmt"
	"log"
	"strings"
	"sync"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/tidwall/gjson"
)

type Exporter struct {
	registry *prometheus.Registry
	metrics  sync.Map
}

func NewExporter() *Exporter {

	return &Exporter{
		registry: prometheus.NewRegistry(),
	}
}

func (e *Exporter) AddMetric(path string, metric string, help string) error {
	_, ok := e.metrics.Load(path)
	if ok {
		log.Printf("error: AddMetric called with duplicate path: %q", path)
		return nil
	}

	parts := strings.SplitN(metric, "_", 3)
	namespace := parts[0]
	subsystem := ""
	name := ""
	if len(parts) > 1 {
		subsystem = parts[1]
	}
	if len(parts) > 2 {
		name = parts[2]
	}

	log.Printf("new gauge: %q -> %q", path, metric)
	pe := prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: namespace,
		Subsystem: subsystem,
		Name:      name,
		Help:      fmt.Sprintf("%s (from %s)", help, path),
	})

	err := e.registry.Register(pe)
	if err != nil {
		log.Printf("error creating new Gauge metric for %q: %v", path, err)
		return err
	}
	e.metrics.Store(path, &pe)
	return nil
}

func (e *Exporter) Export(json string) {
	// Gather all the paths, create all the metrics.
	paths := []string{}
	e.metrics.Range(func(key, value any) bool {
		paths = append(paths, key.(string))
		return true
	})

	// Get all the values, set the gauges.
	values := gjson.GetMany(json, paths...)
	for i, value := range values {
		path := paths[i]
		gaugev, ok := e.metrics.Load(path)
		if !ok {
			log.Printf("error loading gauge for %q", path)
			continue
		}
		gauge := (gaugev).(*prometheus.Gauge)
		(*gauge).Set(value.Num)
		log.Printf("metric: %q = %0.2f", path, value.Num)
	}
}

func (e *Exporter) GetGauge(path string) *prometheus.Gauge {
	ret, _ := e.metrics.Load(path)
	return ret.(*prometheus.Gauge)
}

func (e *Exporter) Close() {
	e.metrics.Range(func(key, value any) bool {
		sink := value.(prometheus.Gauge)
		e.registry.Unregister(sink)
		return true
	})
}
