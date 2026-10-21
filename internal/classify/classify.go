package classify

import (
	"encoding/json"
	"errors"
	"strings"
)

type Metadata struct {
	Title    string   `json:"title"`
	Summary  string   `json:"summary"`
	Category string   `json:"category"`
	Tags     []string `json:"tags"`
	Entities []string `json:"entities"`
	Language string   `json:"language"`
}

var ErrInvalid = errors.New("invalid metadata JSON")

// Parse extracts the first JSON object from a model response and coerces it
// into Metadata. Returns ErrInvalid if no usable object is found.
func Parse(raw string) (Metadata, error) {
	start := strings.IndexByte(raw, '{')
	if start < 0 {
		return Metadata{}, ErrInvalid
	}

	dec := json.NewDecoder(strings.NewReader(raw[start:]))
	dec.UseNumber()

	var m Metadata
	if err := dec.Decode(&m); err != nil {
		return Metadata{}, ErrInvalid
	}

	if m.Title == "" && m.Summary == "" && m.Category == "" {
		return Metadata{}, ErrInvalid
	}

	return m, nil
}
