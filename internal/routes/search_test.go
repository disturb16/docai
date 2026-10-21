package routes

import (
	"testing"

	"github.com/pocketbase/pocketbase/core"
)

func TestMatch(t *testing.T) {
	rtn := core.NewRecord(core.NewBaseCollection("documents"))
	rtn.Set("filename", "rtn_juan_ramos.pdf")
	rtn.Set("title", "RTN Juan Ramos")
	rtn.Set("summary", "Resumen")
	rtn.Set("category", "Legal")
	rtn.Set("tags", []string{"contrato"})
	rtn.Set("entities", []string{"Juan Ramos"})
	rtn.Set("ocr_text", "El contrato fue firmado por Juan.")

	global := core.NewRecord(core.NewBaseCollection("documents"))
	global.Set("filename", "global_contacts.pdf")
	global.Set("title", "Global contacts integration")
	global.Set("summary", "Partners and vendors")
	global.Set("category", "Business")
	global.Set("tags", []string{})
	global.Set("entities", []string{})
	global.Set("ocr_text", "List of partners worldwide.")

	cases := []struct {
		q     string
		rec   *core.Record
		match bool
	}{
		{"rtn juan", rtn, true},
		{"rtn", rtn, true},
		{"rtn", global, false},
		{"juan", rtn, true},
		{"partners", global, true},
		{"rtn partners", global, false},
	}
	for _, c := range cases {
		if got := allIn(tokenize(c.q), haystack(c.rec)); got != c.match {
			t.Errorf("q=%q filename=%q: got %v, want %v", c.q, c.rec.GetString("filename"), got, c.match)
		}
	}
}

func TestToVector(t *testing.T) {
	if got := toVector([]any{1.5, -2.0}); len(got) != 2 || got[0] != 1.5 || got[1] != -2 {
		t.Fatalf("got %v", got)
	}
	if got := toVector([]float32{1, 2}); len(got) != 2 {
		t.Fatalf("got %v", got)
	}
	if got := toVector([]any{1, "x"}); got != nil {
		t.Fatalf("non-numeric input should yield nil, got %v", got)
	}
	if got := toVector(nil); got != nil {
		t.Fatalf("nil should yield nil, got %v", got)
	}
}

func TestCosine(t *testing.T) {
	if got := cosine([]float32{1, 0}, []float32{1, 0}); got != 1 {
		t.Fatalf("got %v, want 1", got)
	}
	if got := cosine([]float32{1, 0}, []float32{0, 1}); got != 0 {
		t.Fatalf("got %v, want 0", got)
	}
}
