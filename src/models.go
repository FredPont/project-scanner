package src

import (
	"encoding/json"
	"fmt"
)

// ProjectReadme holds the raw JSON of a _readme.json file as a generic map,
// so any field — including ones added later in config.json — can be retrieved
// without modifying Go source code.
type ProjectReadme struct {
	raw map[string]interface{}
}

// UnmarshalJSON implements json.Unmarshaler, storing every field in the raw map.
func (r *ProjectReadme) UnmarshalJSON(data []byte) error {
	r.raw = make(map[string]interface{})
	return json.Unmarshal(data, &r.raw)
}

// Project bundles the parsed readme with the filesystem path where it was found.
type Project struct {
	Path   string
	Readme ProjectReadme
}

// GetField returns the string value of any field by its JSON key name.
// Booleans are converted to "true" / "false"; numbers via fmt.Sprint.
// Returns "" if the key is absent.
func (r ProjectReadme) GetField(jsonKey string) string {
	v, ok := r.raw[jsonKey]
	if !ok || v == nil {
		return ""
	}
	switch val := v.(type) {
	case bool:
		if val {
			return "true"
		}
		return "false"
	case string:
		return val
	default:
		return fmt.Sprint(val)
	}
}
