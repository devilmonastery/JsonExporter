package jsonexporter

import (
	"io"
	"log"
	"sync"
	"time"

	"github.com/hashicorp/go-retryablehttp"
	"github.com/prometheus/client_golang/prometheus"
)

var (
	scraperFetches = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: "jsonexporter",
		Subsystem: "scraper",
		Name:      "fetches",
		Help:      "JsonExporter internal scraper fetches",
	}, []string{"url"})
	scraperErrors = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: "jsonexporter",
		Subsystem: "scraper",
		Name:      "errors",
		Help:      "JsonExporter internal scraper errors",
	}, []string{"url"})
)

func init() {
	prometheus.MustRegister(scraperFetches)
	prometheus.MustRegister(scraperErrors)
}

type Scraper struct {
	url      string
	conn     *retryablehttp.Client
	headers  map[string]string
	interval time.Duration
	mu       sync.Mutex
	subs     []chan string
	last     string
	done     chan bool
}

func NewScraper(url string) *Scraper {
	httpcon := retryablehttp.NewClient()
	httpcon.RetryMax = 1
	httpcon.RetryWaitMax = time.Second * 30

	ret := &Scraper{
		url:  url,
		conn: httpcon,
		// standard headers for every request
		headers:  make(map[string]string),
		interval: time.Second * 30,
		done:     make(chan bool),
	}
	go ret.poll()
	return ret
}

func (c *Scraper) Close() {
	close(c.done)
}

func (c *Scraper) poll() {
	log.Printf("starting poll of %s", c.url)
	var wait time.Duration
	for {
		select {
		case <-c.done:
			return
		case <-time.After(wait):
			start := time.Now()
			data, err := c.Get()
			elapsed := time.Since(start)
			wait = c.GetInterval() - elapsed
			if wait < 0 {
				wait = 0
			}
			if err != nil {
				continue
			}
			c.mu.Lock()
			c.last = data
			// broadcast
			for _, s := range c.subs {
				select {
				case s <- data:
				default:
					log.Println("subscriber channel is full")
				}
			}
			c.mu.Unlock()
		}
	}
}

func (c *Scraper) SetInterval(interval time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.interval = interval
}

func (c *Scraper) GetInterval() time.Duration {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.interval
}

func (c *Scraper) SetTimeout(timeout time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.conn.RetryWaitMax = timeout
}

func (c *Scraper) SetRetries(retries uint) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.conn.RetryMax = int(retries)
}

func (c *Scraper) SetHeader(key, value string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.headers[key] = value
}

func (c *Scraper) Get() (ret string, err error) {
	scraperFetches.WithLabelValues(c.url).Inc()
	r, err := retryablehttp.NewRequest("GET", c.url, nil)
	if err != nil {
		scraperErrors.WithLabelValues(c.url).Inc()
		return
	}

	c.mu.Lock()
	for k, v := range c.headers {
		r.Header.Add(k, v)
	}
	c.mu.Unlock()

	res, err := c.conn.Do(r)
	if err != nil {
		scraperErrors.WithLabelValues(c.url).Inc()
		return
	}
	defer res.Body.Close()

	bodybytes, err := io.ReadAll(res.Body)
	if err != nil {
		scraperErrors.WithLabelValues(c.url).Inc()
		return
	}
	ret = string(bodybytes)
	c.mu.Lock()
	c.last = ret
	c.mu.Unlock()
	return
}

func (c *Scraper) Subscribe() chan string {
	ret := make(chan string, 1)
	c.mu.Lock()
	defer c.mu.Unlock()
	c.subs = append(c.subs, ret)
	if c.last != "" {
		ret <- c.last
	}
	return ret
}
