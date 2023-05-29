package jsonexporter

import (
	"testing"
)

func TestLoadConfig(t *testing.T) {
	loader, err := NewConfigLoader("testdata/config.yaml")
	if err != nil {
		t.Fatalf("error loading config: %v", err)
	}
	if loader == nil {
		t.Fatalf("error creating config loader")
	}

	conf := loader.Config()
	if conf == nil {
		t.Fatal("expected config, git nil")
	}

	if len(conf.Targets) == 0 {
		t.Fatalf("expected at least one target, got zero")
	}
}
