package src

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

const readmeFilename = "_readme.json"

// scanProjects walks the root directory up to maxDepth levels deep
// and collects all _readme.json files it can parse successfully.
func ScanProjects(root string, maxDepth int) ([]Project, []string) {
	var projects []Project
	var warnings []string

	rootDepth := strings.Count(filepath.Clean(root), string(os.PathSeparator))

	err := filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			warnings = append(warnings, fmt.Sprintf("⚠ cannot access %s: %v", path, err))
			return nil
		}

		if d.IsDir() {
			currentDepth := strings.Count(filepath.Clean(path), string(os.PathSeparator)) - rootDepth
			if currentDepth > maxDepth {
				return filepath.SkipDir
			}
			return nil
		}

		if d.Name() != readmeFilename {
			return nil
		}

		data, readErr := os.ReadFile(path)
		if readErr != nil {
			warnings = append(warnings, fmt.Sprintf("⚠ cannot read %s: %v", path, readErr))
			return nil
		}

		var readme ProjectReadme
		if jsonErr := json.Unmarshal(data, &readme); jsonErr != nil {
			warnings = append(warnings, fmt.Sprintf("⚠ invalid JSON in %s: %v", path, jsonErr))
			return nil
		}

		projects = append(projects, Project{
			Path:   filepath.Dir(path),
			Readme: readme,
		})
		return nil
	})

	if err != nil {
		warnings = append(warnings, fmt.Sprintf("⚠ walk error: %v", err))
	}

	sort.Slice(projects, func(i, j int) bool {
		return projects[i].Path < projects[j].Path
	})

	return projects, warnings
}
