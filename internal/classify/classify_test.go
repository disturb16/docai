package classify

import "testing"

func TestParse(t *testing.T) {
	cases := []struct {
		name string
		raw  string
		want Metadata
	}{
		{
			name: "clean json",
			raw:  `{"title":"Report","summary":"A report","category":"Finance","tags":["a","b"],"entities":["Acme"],"language":"en"}`,
			want: Metadata{Title: "Report", Summary: "A report", Category: "Finance", Tags: []string{"a", "b"}, Entities: []string{"Acme"}, Language: "en"},
		},
		{
			name: "fenced with prose",
			raw:  "Here you go:\n```json\n{\"title\":\"Notes\",\"summary\":\"s\",\"category\":\"c\",\"tags\":[],\"entities\":[],\"language\":\"en\"}\n```",
			want: Metadata{Title: "Notes", Summary: "s", Category: "c", Language: "en"},
		},
		{
			name: "trailing garbage",
			raw:  `{"title":"T","summary":"s","category":"c"} extra text`,
			want: Metadata{Title: "T", Summary: "s", Category: "c"},
		},
		{
			name: "no json",
			raw:  "I cannot process this document.",
			want: Metadata{},
		},
		{
			name: "empty object",
			raw:  `{}`,
			want: Metadata{},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := Parse(tc.raw)
			if tc.want.Title == "" && tc.want.Summary == "" && tc.want.Category == "" {
				if err == nil {
					t.Fatalf("expected error, got %+v", got)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got.Title != tc.want.Title || got.Summary != tc.want.Summary || got.Category != tc.want.Category || got.Language != tc.want.Language {
				t.Fatalf("got %+v, want %+v", got, tc.want)
			}
		})
	}
}
