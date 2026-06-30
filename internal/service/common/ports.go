// Package common defines cross-cutting Port interfaces shared by multiple usecases and infra.
package common

import "context"

// SearchOpts are shared search options (used by docs/maven/log search).
type SearchOpts struct {
	Literal    bool
	IgnoreCase bool
	Glob       string
	Type       string
	Context    ContextLines
	MaxCount   int
	Raw        bool
}

// ContextLines configures surrounding lines (spec 6.4.1: context {A,B,C}).
type ContextLines struct {
	A int // before
	B int // after
	C int // both (shorthand for A=B=C)
}

// Match is a single search hit.
type Match struct {
	File        string `json:"file"`
	Line        int    `json:"line"`
	LineContent string `json:"lineContent"`
	Type        string `json:"type"` // "match" | "context"
}

// SearchResult is the truncated (non-paged) search response.
type SearchResult struct {
	Items     []Match `json:"items"`
	Total     int     `json:"total"`
	Truncated bool    `json:"truncated"`
}

// Searcher is the Port for full-text search (implemented by infra/searcher).
type Searcher interface {
	Search(ctx context.Context, roots []string, query string, opts SearchOpts) (*SearchResult, error)
}
