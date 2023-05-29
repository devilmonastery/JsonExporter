package jsonexporter

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestScraper(t *testing.T) {
	testval := "this is a test"

	svr := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, testval)
	}))
	defer svr.Close()

	s := NewScraper(svr.URL)
	s.SetInterval(time.Second)
	defer s.Close()

	val, err := s.Get()
	if err != nil {
		t.Fatalf("failed to get data")
	}

	if val != testval {
		t.Errorf("expected %q, got %q", testval, val)
	}

	ch := s.Subscribe()

	val = <-ch
	if val != testval {
		t.Errorf("expected %q, got %q", testval, val)
	}

}
