package jsonexporter

import (
	"time"

	"github.com/devilmonastery/configloader"
)

type Config struct {
	Targets []Target `yaml:"targets"`
}

type Target struct {
	Name     string        `yaml:"name"`
	URL      string        `yaml:"url"`      // endpoint to poll
	Interval time.Duration `yaml:"interval"` // how often to poll
	Metrics  []Metric      `yaml:"metrics"`
}

type Metric struct {
	Path string `yaml:"path"` // json path expression
	Name string `yaml:"name"` // namespace_subsystem_name
	Help string `yaml:"help"` // metric description
}

func NewConfigLoader(path string) (*configloader.ConfigLoader[Config], error) {
	return configloader.NewConfigLoader[Config](path)
}
