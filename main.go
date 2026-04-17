/*
 This program is free software: you can redistribute it and/or modify
 it under the terms of the GNU General Public License as published by
 the Free Software Foundation, either version 3 of the License, or
 (at your option) any later version.

 This program is distributed in the hope that it will be useful,
 but WITHOUT ANY WARRANTY; without even the implied warranty of
 MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 GNU General Public License for more details.

 You should have received a copy of the GNU General Public License
 along with this program.  If not, see <http://www.gnu.org/licenses/>.

 Written by Frederic PONT.
 (c) Frederic Pont 2026
*/

package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"project-scanner/src"
)

func main() {
	src.EnableANSI()                // activate ANSI colours on Windows
	defer src.WaitIfDoubleClicked() // pause if double-clicked
	const version = "2026-04-16"    // software version

	// ── Pré-lire -config et -filename sans flag.Parse ─────────────────────
	// We need the config path before registering flags, so we scan os.Args
	// manually for -config / --config to avoid a chicken-and-egg problem.
	configPath := findArg(os.Args[1:], "config", "config.json")
	filename := findArg(os.Args[1:], "filename", "_readme.json")

	// ── Load config early ─────────────────────────────────────────────────
	cfg, err := src.LoadConfig(configPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "❌ %v\n", err)
		os.Exit(1)
	}

	// ── Register ALL static flags ─────────────────────────────────────────
	root := flag.String("root", ".", "Root directory to scan")
	maxDepth := flag.Int("depth", 5, "Maximum folder depth (0 = root only)")
	flag.String("config", configPath, "Path to the JSON configuration file")
	flag.String("filename", filename, "Name of the readme JSON file to look for")
	showVersion := flag.Bool("v", false, "Print version and exit")

	// ── Register dynamic flags from config ────────────────────────────────────
	// For each filterable field we register one or two flags:
	//   - date_range → -<slug>-before  and  -<slug>-after
	//   - exact / contains → -<slug>
	//
	// "slug" is the JSON key lowercased with underscores replaced by hyphens.

	type dynFlag struct {
		field    src.FieldConfig
		before   *string
		after    *string
		exact    *string
		contains *string
	}
	var dynFlags []dynFlag

	for _, f := range cfg.FilterableFields() {
		slug := strings.ToLower(strings.ReplaceAll(f.JSONKey, "_", "-"))
		df := dynFlag{field: f}

		switch f.Filter {
		case src.FilterDateRange:
			df.before = flag.String(slug+"-before", "", fmt.Sprintf("Keep projects whose %s is before YYYY-MM-DD", f.JSONKey))
			df.after = flag.String(slug+"-after", "", fmt.Sprintf("Keep projects whose %s is after YYYY-MM-DD", f.JSONKey))
		case src.FilterExact:
			df.exact = flag.String(slug, "", fmt.Sprintf("Filter by %s (exact match, case-insensitive)", f.JSONKey))
		case src.FilterContains:
			df.contains = flag.String(slug, "", fmt.Sprintf("Filter by %s (substring match, case-insensitive)", f.JSONKey))
		}
		dynFlags = append(dynFlags, df)
	}

	// ── Custom usage (registered before second parse so -h shows everything) ──
	flag.Usage = func() {
		fmt.Fprint(os.Stderr, `
╔══════════════════════════════════════════════════════════════╗
║   Project Scanner v`+version+` - ©Frédéric Pont - GNU GPL     ║
╚══════════════════════════════════════════════════════════════╝

Usage: project-scanner [options]

Options:
`)
		flag.PrintDefaults()
		fmt.Fprint(os.Stderr, `
Examples:
  project-scanner -root ./data
  project-scanner -root ./data -depth 3 -config myconfig.json
  project-scanner -root ./data -project-status ongoing
  project-scanner -root ./data -project-published false
  project-scanner -root ./data -end-date-before 2025-01-01
  project-scanner -root ./data -end-date-after 2024-01-01
`)
	}

	// ── Single parse — all flags known ────────────────────────────────────
	flag.Parse()

	// Version flag — handled before anything else
	if *showVersion {
		fmt.Printf("project-scanner version %s\n", version)
		return
	}

	// // ── Second parse: all flags (static + dynamic) ────────────────────────────
	// if err := flag.CommandLine.Parse(os.Args[1:]); err != nil {
	// 	if errors.Is(err, flag.ErrHelp) {
	// 		flag.Usage()
	// 		os.Exit(0)
	// 	}
	// 	// gestion of other errors (ex: bad type of flag)
	// 	fmt.Fprintf(os.Stderr, "Erreur: %v\n", err)
	// 	os.Exit(2)
	// }

	// ── Build FilterSet ───────────────────────────────────────────────────────
	var filterSet src.FilterSet
	for _, df := range dynFlags {
		ff := src.FieldFilter{JSONKey: df.field.JSONKey}
		switch df.field.Filter {
		case src.FilterDateRange:
			ff.Before = *df.before
			ff.After = *df.after
		case src.FilterExact:
			ff.Exact = *df.exact
		case src.FilterContains:
			ff.Contains = *df.contains
		}
		if ff.Before != "" || ff.After != "" || ff.Exact != "" || ff.Contains != "" {
			filterSet = append(filterSet, ff)
		}
	}

	// ── Scan ──────────────────────────────────────────────────────────────────
	fmt.Printf("\033[36;1m🔍\033[0m Scanning \033[1m%s\033[0m (max depth: %d)…\n", *root, *maxDepth)

	projects, warnings := src.ScanProjects(*root, *maxDepth, filename)
	for _, w := range warnings {
		fmt.Printf("\033[33m%s\033[0m\n", w)
	}
	fmt.Printf("\033[36m📁\033[0m Found \033[1m%d\033[0m project(s)\n", len(projects))

	// ── Filter ────────────────────────────────────────────────────────────────
	filtered := src.Apply(projects, filterSet)
	fmt.Printf("\033[36m🎯\033[0m Matching filter(s): \033[1;32m%d\033[0m project(s)\n\n", len(filtered))

	if len(filtered) == 0 {
		fmt.Println("\033[33mNo projects match the current filters.\033[0m")
		return
	}

	// ── Render ────────────────────────────────────────────────────────────────
	src.PrintTable(filtered, cfg)
	fmt.Printf("\n\033[1m%d result(s)\033[0m\n", len(filtered))
}

// findArg scans args for -name=value or -name value, returning defaultVal if absent.
func findArg(args []string, name, defaultVal string) string {
	for i, a := range args {
		// Handle -name=value and --name=value
		for _, prefix := range []string{"-" + name + "=", "--" + name + "="} {
			if strings.HasPrefix(a, prefix) {
				return strings.TrimPrefix(a, prefix)
			}
		}
		// Handle -name value and --name value
		if (a == "-"+name || a == "--"+name) && i+1 < len(args) {
			return args[i+1]
		}
	}
	return defaultVal
}
