package routes

import (
	"encoding/json"
	"time"

	"github.com/pocketbase/pocketbase/core"
)

type analyticsResponse struct {
	Totals    map[string]int       `json:"totals"` // per status: pending, processing, done, error
	Daily     []dailyBucket        `json:"daily"`  // docs created per day, last 14 days
	Providers map[string]providerStats `json:"providers"` // per model (ollama/llamacpp/raw)
}

type dailyBucket struct {
	Date  string `json:"date"`
	Total int    `json:"total"`
	Done  int    `json:"done"`
	Error int    `json:"error"`
}

type providerStats struct {
	Count         int `json:"count"`
	Done          int `json:"done"`
	Error         int `json:"error"`
	AvgTotalMs    int `json:"avg_total_ms"`
	AvgExtractMs  int `json:"avg_extract_ms"`
	AvgOCRMs      int `json:"avg_ocr_ms"`
	AvgClassifyMs int `json:"avg_classify_ms"`
	AvgRetries    int `json:"avg_retries"`
}

// registerAnalytics adds GET /api/docai/analytics with aggregated processing
// stats derived from the documents collection.
func registerAnalytics(app core.App, e *core.ServeEvent) {
	e.Router.GET("/api/docai/analytics", func(e *core.RequestEvent) error {
		// ponytail: in-memory scan of all records; switch to SQL GROUP BY
		// when the library grows past a few thousand docs
		records, err := app.FindAllRecords("documents")
		if err != nil {
			return e.InternalServerError("failed to fetch documents", err)
		}
		return e.JSON(200, aggregateAnalytics(records, time.Now()))
	})
}

func aggregateAnalytics(records []*core.Record, now time.Time) analyticsResponse {
	resp := analyticsResponse{
		Totals:    map[string]int{"pending": 0, "processing": 0, "done": 0, "error": 0},
		Providers: map[string]providerStats{},
	}

	buckets := make(map[string]*dailyBucket, 14)
	for i := 13; i >= 0; i-- {
		d := now.AddDate(0, 0, -i).Format("2006-01-02")
		b := &dailyBucket{Date: d}
		buckets[d] = b
		resp.Daily = append(resp.Daily, *b)
	}

	sums := map[string]map[string]int{}

	for _, r := range records {
		status := r.GetString("status")
		resp.Totals[status]++

		if d := r.GetDateTime("created").Time(); !d.IsZero() {
			if b, ok := buckets[d.Format("2006-01-02")]; ok {
				b.Total++
				if status == "done" {
					b.Done++
				}
				if status == "error" {
					b.Error++
				}
			}
		}

		provider := r.GetString("model")
		if provider == "" {
			provider = "unknown"
		}
		ps := resp.Providers[provider]
		ps.Count++
		if status == "done" {
			ps.Done++
		}
		if status == "error" {
			ps.Error++
		}
		ps.AvgRetries += r.GetInt("retry_count")
		resp.Providers[provider] = ps

		ss, ok := sums[provider]
		if !ok {
			ss = map[string]int{}
			sums[provider] = ss
		}
		if timings, ok := r.Get("timings").(map[string]interface{}); ok {
			total := 0
			for _, v := range timings {
				total += toInt(v)
			}
			ss["total"] += total
			ss["extract"] += toInt(timings["extract"])
			ss["ocr"] += toInt(timings["ocr"])
			ss["classify"] += toInt(timings["classify"])
		}
	}

	for provider, ps := range resp.Providers {
		n := ps.Count
		ss := sums[provider]
		ps.AvgTotalMs = divide(ss["total"], n)
		ps.AvgExtractMs = divide(ss["extract"], n)
		ps.AvgOCRMs = divide(ss["ocr"], n)
		ps.AvgClassifyMs = divide(ss["classify"], n)
		ps.AvgRetries = ps.AvgRetries / n
		resp.Providers[provider] = ps
	}

	return resp
}

func divide(sum, count int) int {
	if count == 0 {
		return 0
	}
	return sum / count
}

func toInt(v interface{}) int {
	switch n := v.(type) {
	case float64:
		return int(n)
	case int:
		return n
	case int64:
		return int(n)
	case json.Number:
		i, _ := n.Int64()
		return int(i)
	}
	return 0
}
