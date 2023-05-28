package json2metrics

import (
	"os"
	"testing"

	"github.com/prometheus/client_golang/prometheus/testutil"
)

func TestExport(t *testing.T) {
	data, err := os.ReadFile("testdata/model.json")
	if err != nil {
		t.Fatalf("could not read test data")
	}

	exp := NewExporter()

	type test struct {
		path   string
		metric string
		value  float64
	}

	tests := []test{
		{path: "sms.unreadMsgs", metric: "test_sms_unread", value: 0.0},
		{path: "general.upTime", metric: "test_general_uptime", value: 2078.0},
	}

	for _, tst := range tests {
		err := exp.AddMetric(tst.path, tst.metric, "test")
		if err != nil {
			t.Fatalf("error adding metric %q (%s): %v", tst.path, tst.metric, err)
		}
	}

	exp.Export(string(data))

	for i, tst := range tests {
		g := exp.GetGauge(tst.path)
		if g == nil {
			t.Errorf("test %d: expected metric for path %q", i, tst.path)
			continue
		}
		val := testutil.ToFloat64(*g)
		if val != tst.value {
			t.Errorf("test %d: expected metric for path %q = %0.2f; got: %0.2f", i, tst.path, tst.value, val)

		}

	}
}
