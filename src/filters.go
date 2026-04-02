package src

import (
	"strings"
	"time"
)

const dateLayout = "2006-01-02" // ISO-8601

// FieldFilter holds the runtime filter values for a single field.
// For date_range fields: Before and/or After may be set.
// For exact fields: Exact may be set.
type FieldFilter struct {
	JSONKey  string
	Before   string // date_range: keep records whose date < Before
	After    string // date_range: keep records whose date > After
	Exact    string // exact: keep records whose value == Exact (case-insensitive)
	Contains string // contains: search for a substring in records
}

// FilterSet is a collection of active field filters.
type FilterSet []FieldFilter

// parseDate parses an ISO date string; returns zero time on failure.
func parseDate(s string) (time.Time, bool) {
	if s == "" {
		return time.Time{}, false
	}
	t, err := time.Parse(dateLayout, s)
	return t, err == nil
}

// Matches returns true when the project satisfies every active filter in the set.
func (fs FilterSet) Matches(p Project) bool {
	for _, f := range fs {
		value := p.Readme.GetField(f.JSONKey)

		// --- date_range ---
		if f.Before != "" {
			threshold, ok := parseDate(f.Before)
			if ok {
				fieldDate, fok := parseDate(value)
				if !fok || !fieldDate.Before(threshold) {
					return false
				}
			}
		}
		if f.After != "" {
			threshold, ok := parseDate(f.After)
			if ok {
				fieldDate, fok := parseDate(value)
				if !fok || !fieldDate.After(threshold) {
					return false
				}
			}
		}

		// --- exact (string / bool) ---
		if f.Exact != "" {
			if !strings.EqualFold(value, f.Exact) {
				return false
			}
		}

		// --- contains (string) ---
		if f.Contains != "" {
			if !strings.Contains(value, f.Contains) {
				return false
			}
		}
	}
	return true
}

// Apply filters a project slice, returning only those that match.
func Apply(projects []Project, fs FilterSet) []Project {
	var out []Project
	for _, p := range projects {
		if fs.Matches(p) {
			out = append(out, p)
		}
	}
	return out
}
