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
	"errors"
	"flag"
	"fmt"
	"os"
	"strings"

	"project-scanner/src"
)

func main() {
	src.EnableANSI()                // activate coulors ANSI on Windows
	defer src.WaitIfDoubleClicked() // pause if double-clic
	const version = "2026-04-15"    // software version

	// ── Static flags ──────────────────────────────────────────────────────────
	root := flag.String("root", ".", "Root directory to scan")
	maxDepth := flag.Int("depth", 5, "Maximum folder depth (0 = root only)")
	configPath := flag.String("config", "config.json", "Path to the JSON configuration file")
	filename := flag.String("filename", "_readme.json", "Name of the readme JSON file to look for")
	showVersion := flag.Bool("version", false, "Print version and exit")
	// ── Load config (before parsing remaining flags) ──────────────────────────
	// We do a pre-parse to get -config if supplied, then load the config file,
	// then register dynamic flags, then do the final parse.
	flag.CommandLine.Init(os.Args[0], flag.ContinueOnError)
	//_ = flag.CommandLine.Parse(os.Args[1:]) // first pass — static flags only

	// Vérify if flag "version" is activated, if yes print version and exit
	// if *showVersion {
	// 	fmt.Printf("Version : %s\n", version)
	// 	//os.Exit(0)
	// 	return
	// }

	cfg, err := src.LoadConfig(*configPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "❌ %v\n", err)
		os.Exit(1)
	}

	// ── Dynamic flags from config ─────────────────────────────────────────────
	// For each filterable field we register one or two flags:
	//   - date_range → -<slug>-before  and  -<slug>-after
	//   - exact      → -<slug>
	//
	// "slug" is the JSON key lowercased with underscores replaced by dashes.

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

	// Custom usage
	flag.Usage = func() {
		fmt.Fprint(os.Stderr, `
╔══════════════════════════════════════════════════════════════╗
║    Project Scanner v`+version+` - ©Fréderic Pont - Gnu GPL     ║
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

	// Second parse — picks up dynamic flags now that they are registered.
	flag.Parse()
	// if arg -h, then exit
	if err := flag.CommandLine.Parse(os.Args[1:]); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			os.Exit(0)
		}
		os.Exit(2)
	}

	if *showVersion {
		fmt.Printf("Version : %s\n", version)
		//os.Exit(0)
		return
	}

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
		// Only add to the set if at least one value is set
		if ff.Before != "" || ff.After != "" || ff.Exact != "" || ff.Contains != "" {
			filterSet = append(filterSet, ff)
		}
	}

	// ── Scan ──────────────────────────────────────────────────────────────────
	fmt.Printf("\033[36;1m🔍\033[0m Scanning \033[1m%s\033[0m (max depth: %d)…\n", *root, *maxDepth)

	projects, warnings := src.ScanProjects(*root, *maxDepth, *filename)
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
