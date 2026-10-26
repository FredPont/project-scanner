package src

import (
	"fmt"
	"regexp"
	"strings"
)

// ─────────────────────────────────────────────
// ANSI colour helpers
// ─────────────────────────────────────────────

const (
	reset  = "\033[0m"
	bold   = "\033[1m"
	red    = "\033[31m"
	green  = "\033[32m"
	yellow = "\033[33m"
	cyan   = "\033[36m"
	white  = "\033[37m"
	bgBlue = "\033[44m"
	bgGrey = "\033[100m"
)

// col wraps a string in ANSI color codes, or returns it unchanged if color is empty.
func col(c, s string) string { return c + s + reset }

// getANSIColor convertit un nom de couleur textuel (provenant du JSON) en code ANSI réel.
func getANSIColor(colorName string) string {
	switch strings.ToLower(colorName) {
	case "green":
		return green
	case "yellow":
		return yellow
	case "cyan":
		return cyan
	case "white":
		return white
	case "red":
		return red
	default:
		return red // Rouge par défaut si non spécifié ou inconnu
	}
}

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

func publishedLabel(pub bool) string {
	if pub {
		return col(green, symOK+"yes")
	}
	return col(red, symNO+"no")
}

func unresponsiveLabel(unresponsive bool) string {
	if unresponsive {
		return col(red, symNO+"yes")
	}
	return col(green, symOK+"no")
}

// ─────────────────────────────────────────────
// Table layout
// ─────────────────────────────────────────────

const pathColWidth = 50

// totalWidth computes the exact printed width of a table row:
// 1 leading space + PATH col + one space per extra col + each col width.
func totalWidth(cols []FieldConfig) int {
	w := 1 + pathColWidth // leading space + PATH col
	for _, c := range cols {
		w += 1 + c.ColWidth // separator space + col content
	}
	return w
}

// visibleLen returns the visible character count of a string,
// ignoring ANSI escape codes. Uses rune count to handle multibyte glyphs.
func visibleLen(s string) int {
	return len([]rune(stripANSI(s)))
}

// tableSeparator prints a full-width separator line matching the table width.
func tableSeparator(cols []FieldConfig) {
	fmt.Println(col(bgGrey, strings.Repeat(" ", totalWidth(cols))))
}

func tableHeader(cols []FieldConfig) {
	// Build the full header line as a single colored string,
	// so no stray reset codes break the spacing between columns.
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("%-*s", pathColWidth, "PATH"))
	for _, c := range cols {
		sb.WriteString(" ")
		sb.WriteString(fmt.Sprintf("%-*s", c.ColWidth, c.Label))
	}

	// Apply color once around the entire padded line
	fmt.Println(bold + bgBlue + white + " " + sb.String() + reset)
}

func tableRow(p Project, cols []FieldConfig, even bool) {
	bg := ""
	if !even {
		bg = "\033[48;5;236m"
	}

	// PATH column: always plain text, no color
	path := fmt.Sprintf("%-*s", pathColWidth, truncate(p.Path, pathColWidth))
	fmt.Printf("%s %s", bg, path)

	for _, c := range cols {
		raw := p.Readme.GetField(c.JSONKey)

		var cell string
		switch c.Render {
		case "status":
			// Rendu piloté dynamiquement par le JSON (color_map)
			padded := fmt.Sprintf("%-*s", c.ColWidth, truncate(raw, c.ColWidth))

			colorCode := red // Couleur par défaut si la valeur n'est pas dans la map
			if c.ColorMap != nil {
				if colName, ok := c.ColorMap[strings.ToLower(raw)]; ok {
					colorCode = getANSIColor(colName)
				}
			}
			cell = colorCode + padded + reset + bg

		case "boolean":
			label := publishedLabel(raw == "true")
			padding := c.ColWidth - visibleLen(label)
			if padding < 0 {
				padding = 0
			}
			cell = label + bg + strings.Repeat(" ", padding) + reset

		default:
			// Plain date/string: safe to use %-*s directly
			cell = fmt.Sprintf("%-*s", c.ColWidth, truncate(raw, c.ColWidth))
		}

		fmt.Printf(" %s", cell)
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

var ansiRe = regexp.MustCompile(`\x1b\[[0-9;]*m`)

// stripANSI removes ANSI escape codes to get the real visible length of a string.
func stripANSI(s string) string {
	return ansiRe.ReplaceAllString(s, "")
}
