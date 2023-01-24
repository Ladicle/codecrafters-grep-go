package grep

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRun(t *testing.T) {
	var tests = []struct {
		name        string
		s           string
		pattern     string
		wantMatched bool
	}{
		{
			name:        "match literal character",
			s:           "apple",
			pattern:     "a",
			wantMatched: true,
		},
		{
			name:        "unmatch literal character",
			s:           "dog",
			pattern:     "a",
			wantMatched: false,
		},
		{
			name:        "match digits",
			s:           "apple123",
			pattern:     "\\d",
			wantMatched: true,
		},
		{
			name:        "unmatch digits",
			s:           "apple",
			pattern:     "\\d",
			wantMatched: false,
		},
		{
			name:        "match alphanumeric characters",
			s:           "alpha-num3ric",
			pattern:     "\\w",
			wantMatched: true,
		},
		{
			name:        "unmatch alphanumeric characters",
			s:           ".....",
			pattern:     "\\w",
			wantMatched: false,
		},
		{
			name:        "match positive character groups",
			s:           "apple",
			pattern:     "[abc]",
			wantMatched: true,
		},
		{
			name:        "unmatch positive character groups",
			s:           "dog",
			pattern:     "[abc]",
			wantMatched: false,
		},
		{
			name:        "match negative character groups",
			s:           "dog",
			pattern:     "[^abc]",
			wantMatched: true,
		},
		{
			name:        "unmatch negative character groups",
			s:           "apple",
			pattern:     "[^abc]",
			wantMatched: false,
		},
		{
			name:        "match combining character classes",
			s:           "10 apple",
			pattern:     `\d\d ap\wle`,
			wantMatched: true,
		},
		{
			name:        "unmatch combining character classes",
			s:           "1 dog",
			pattern:     `\d \w\w\ws`,
			wantMatched: false,
		},
		{
			name:        "match start of string anchor",
			s:           "log",
			pattern:     "^log",
			wantMatched: true,
		},
		{
			name:        "unmatch start of string anchor",
			s:           "slog",
			pattern:     "^log",
			wantMatched: false,
		},
		{
			name:        "match end of string anchor",
			s:           "dog",
			pattern:     "dog$",
			wantMatched: true,
		},
		{
			name:        "unmatch end of string anchor",
			s:           "dogs",
			pattern:     "dog$",
			wantMatched: false,
		},
		{
			name:        "match one or more times",
			s:           "caaaaaaaaaaats",
			pattern:     "ca+ts",
			wantMatched: true,
		},
		{
			name:        "match one times",
			s:           "cats",
			pattern:     "ca+ts",
			wantMatched: true,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			matched := Run(test.s, test.pattern)
			assert.Equal(t, test.wantMatched, matched)
		})
	}
}
