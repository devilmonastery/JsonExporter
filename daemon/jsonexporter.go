package main

import (
	"flag"
	"log"
	"net/http"

	"devilmonastery/jsonexporter"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	listen   = flag.String("listen", ":9421", "Address on which to serve")
	confFile = flag.String("config", "config.yaml", "path to the config file")

	errors = prometheus.NewCounter(prometheus.CounterOpts{
		Namespace: "jsonexporter",
		Subsystem: "exporter",
		Name:      "errors",
		Help:      "JsonExporter internal errors",
	})
)

func init() {
	prometheus.MustRegister(errors)
}

type Exp struct {
	t    jsonexporter.Target
	s    *jsonexporter.Scraper
	e    *jsonexporter.Exporter
	done chan bool
}

func NewExp(t jsonexporter.Target) *Exp {
	ret := &Exp{
		t:    t,
		s:    jsonexporter.NewScraper(t.URL),
		e:    jsonexporter.NewExporter(),
		done: make(chan bool),
	}
	if ret.t.Interval > 0 {
		ret.s.SetInterval(ret.t.Interval)
	}
	for _, m := range t.Metrics {
		err := ret.e.AddMetric(m.Path, m.Name, m.Help)
		if err != nil {
			log.Fatalf("error adding metric %q: %v", m.Path, err)
		}
	}
	go ret.Run()
	return ret
}

func (c *Exp) Run() {
	ch := c.s.Subscribe()
	for {
		select {
		case <-c.done:
			log.Printf("exiting exporter")
			return
		case data := <-ch:
			log.Printf("got data for %s", c.t.URL)
			c.e.Export(data)
		}
	}
}

func (c *Exp) Close() {
	c.s.Close()
	c.e.Close()
	close(c.done)
}

func loop(confs chan jsonexporter.Config) {
	exporters := []*Exp{}

	for {
		select {
		case conf := <-confs:
			log.Printf("new config, resetting exporters")
			for _, e := range exporters {
				e.Close()
			}
			exporters = []*Exp{}
			for _, t := range conf.Targets {
				e := NewExp(t)
				exporters = append(exporters, e)
			}
		}
	}

}

func main() {
	flag.Parse()
	log.SetFlags(log.Lshortfile)
	http.Handle("/metrics", promhttp.Handler())

	loader, err := jsonexporter.NewConfigLoader(*confFile)
	if err != nil {
		log.Fatalf("error reading config at %q: %v", *confFile, err)
	}

	go loop(loader.Subscribe())

	log.Print(http.ListenAndServe(*listen, nil))
}
