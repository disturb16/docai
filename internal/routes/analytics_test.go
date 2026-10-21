package routes

import (
	"testing"
	"time"

	"github.com/pocketbase/pocketbase/core"
)

func TestAggregateAnalytics(t *testing.T) {
	mk := func(status, model string, retries int, timings map[string]interface{}) *core.Record {
		r := core.NewRecord(core.NewBaseCollection("documents"))
		r.Set("status", status)
		r.Set("model", model)
		r.Set("retry_count", retries)
		if timings != nil {
			r.Set("timings", timings)
		}
		return r
	}

	records := []*core.Record{
		mk("done", "ollama", 0, map[string]interface{}{"extract": 100.0, "ocr": 200.0, "classify": 300.0}),
		mk("done", "ollama", 0, map[string]interface{}{"extract": 100.0, "ocr": 200.0, "classify": 300.0}),
		mk("error", "llamacpp", 3, nil),
	}

	now := time.Date(2026, 9, 4, 12, 0, 0, 0, time.UTC)
	got := aggregateAnalytics(records, now)

	if got.Totals["done"] != 2 || got.Totals["error"] != 1 || got.Totals["pending"] != 0 {
		t.Errorf("totals wrong: %+v", got.Totals)
	}

	ollama := got.Providers["ollama"]
	if ollama.Count != 2 || ollama.Done != 2 || ollama.AvgClassifyMs != 300 || ollama.AvgTotalMs != 600 {
		t.Errorf("ollama stats wrong: %+v", ollama)
	}

	llama := got.Providers["llamacpp"]
	if llama.Count != 1 || llama.Error != 1 || llama.AvgRetries != 3 || llama.AvgTotalMs != 0 {
		t.Errorf("llamacpp stats wrong: %+v", llama)
	}

	if len(got.Daily) != 14 {
		t.Errorf("expected 14 daily buckets, got %d", len(got.Daily))
	}
}
