package src

import (
	"fmt"
	"strings"
)

// ─────────────────────────────────────────────
// ANSI colour helpers
// ─────────────────────────────────────────────

const (
	reset   = "\033[0m"
	bold    = "\033[1m"
	red     = "\033[31m"
	green   = "\033[32m"
	yellow  = "\033[33m"
	cyan    = "\033[36m"
	white   = "\033[37m"
	bgBlue  = "\033[44m"
	bgGrey  = "\033[100m"
)

func col(c, s string) string { return c + s + reset }

// ─────────────────────────────────────────────
// Field rendering helpers
// ─────────────────────────────────────────────

// truncate shortens a string to maxLen characters, adding "…" if needed.
func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-1] + "…"
}

// renderValue applies visual formatting depending on the field type/key.
func renderValue(cfg FieldConfig, raw string) string {
	switch cfg.JSONKey {
	case "Project_status":
		return statusColored(raw, cfg.ColWidth)
	case "Project_published":
		return publishedLabel(raw == "true")
	default:
		return truncate(raw, cfg.ColWidth)
	}
}

func statusColored(status string, width int) string {
	var color string
	switch strings.ToLower(status) {
	case "ongoing":
		color = yellow
	case "completed":
		color = green
	default:
		color = red
	}
	label := truncate(status, width)
	return col(color, label)
}

func publishedLabel(pub bool) string {
	if pub {
		return col(green, "✔ yes")
	}
	return col(red, "✘ no")
}

// ─────────────────────────────────────────────
// Table layout
// ─────────────────────────────────────────────

const pathColWidth = 50

// totalWidth computes the full table width from the config.
func totalWidth(cols []FieldConfig) int {
	w := 1 + pathColWidth + 1 // leading space + PATH col + space
	for _, c := range cols {
		w += c.ColWidth + 1
	}
	return w
}

func tableSeparator(cols []FieldConfig) {
	fmt.Println(col(bgGrey, strings.Repeat(" ", totalWidth(cols))))
}

func tableHeader(cols []FieldConfig) {
	fmt.Print(col(bold+bgBlue, " "))
	fmt.Print(col(bold+bgBlue+white, fmt.Sprintf(" %-*s", pathColWidth, "PATH")))
	for _, c := range cols {
		fmt.Print(col(bold+bgBlue+white, fmt.Sprintf(" %-*s", c.ColWidth, c.Label)))
	}
	fmt.Println(reset)
}

func tableRow(p Project, cols []FieldConfig, even bool) {
	bg := ""
	if !even {
		bg = "\033[48;5;236m"
	}

	path := truncate(p.Path, pathColWidth)
	fmt.Printf("%s %-*s", bg, pathColWidth, path)

	for _, c := range cols {
		raw := p.Readme.GetField(c.JSONKey)
		rendered := renderValue(c, raw)
		// For plain fields, pad to col width; colored fields carry their own ANSI codes.
		switch c.JSONKey {
		case "Project_status", "Project_published":
			fmt.Printf(" %s%s", rendered, bg)
		default:
			fmt.Printf(" %-*s", c.ColWidth, rendered)
		}
	}
	fmt.Printf("%s\n", reset)
}

// PrintTable renders the full coloured table for a list of projects,
// using only the columns declared as ShowInTable in the config.
func PrintTable(projects []Project, cfg AppConfig) {
	cols := cfg.TableFields()
	tableSeparator(cols)
	tableHeader(cols)
	tableSeparator(cols)
	for i, p := range projects {
		tableRow(p, cols, i%2 == 0)
	}
	tableSeparator(cols)
}
