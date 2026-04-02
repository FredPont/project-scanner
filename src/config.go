package src

import (
	"encoding/json"
	"fmt"
	"os"
)

// FieldType describes the kind of value stored in a _readme.json field.
type FieldType string

const (
	FieldTypeDate   FieldType = "date"
	FieldTypeString FieldType = "string"
	FieldTypeBool   FieldType = "bool"
)

// FilterMode describes how a field can be filtered from the CLI.
type FilterMode string

const (
	FilterNone      FilterMode = "none"       // not filterable
	FilterExact     FilterMode = "exact"      // --field <value>  (string / bool equality)
	FilterDateRange FilterMode = "date_range" // --field-before / --field-after
	FilterContains  FilterMode = "contains"   // --field <value>  (substring equality)
)

// FieldConfig describes one field of _readme.json, controlling both filtering
// and display behaviour.
type FieldConfig struct {
	// JSONKey is the exact key used in _readme.json (e.g. "End_Date").
	JSONKey string `json:"json_key"`

	// Label is the column header shown in the table.
	Label string `json:"label"`

	// Type is the data type: "date" | "string" | "bool".
	Type FieldType `json:"type"`

	// Filter controls how this field can be filtered from the CLI.
	// "none" | "exact" | "date_range"
	Filter FilterMode `json:"filter"`

	// ShowInTable controls whether this field appears as a column.
	ShowInTable bool `json:"show_in_table"`

	// ColWidth is the display width (in characters) for this column.
	// Used only when ShowInTable is true.
	ColWidth int `json:"col_width"`
}

// AppConfig is the root structure of config.json.
type AppConfig struct {
	// Fields defines all known _readme.json fields and how to handle them.
	Fields []FieldConfig `json:"fields"`
}

// DefaultConfig returns a sensible default configuration that reproduces
// the original hard-coded behaviour.
func DefaultConfig() AppConfig {
	return AppConfig{
		Fields: []FieldConfig{
			{
				JSONKey:     "Start_Date",
				Label:       "START",
				Type:        FieldTypeDate,
				Filter:      FilterNone,
				ShowInTable: true,
				ColWidth:    12,
			},
			{
				JSONKey:     "End_Date",
				Label:       "END",
				Type:        FieldTypeDate,
				Filter:      FilterDateRange,
				ShowInTable: true,
				ColWidth:    12,
			},
			{
				JSONKey:     "All_Data_transfered",
				Label:       "TRANSFER",
				Type:        FieldTypeDate,
				Filter:      FilterDateRange,
				ShowInTable: true,
				ColWidth:    12,
			},
			{
				JSONKey:     "Project_status",
				Label:       "STATUS",
				Type:        FieldTypeString,
				Filter:      FilterExact,
				ShowInTable: true,
				ColWidth:    12,
			},
			{
				JSONKey:     "Project_published",
				Label:       "PUBLISHED",
				Type:        FieldTypeBool,
				Filter:      FilterExact,
				ShowInTable: true,
				ColWidth:    10,
			},
		},
	}
}

// LoadConfig reads and parses a config.json file from disk.
// If the file does not exist, it writes the default config to disk and returns it.
func LoadConfig(path string) (AppConfig, error) {
	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		cfg := DefaultConfig()
		if writeErr := WriteConfig(path, cfg); writeErr != nil {
			fmt.Fprintf(os.Stderr, "⚠ could not write default config: %v\n", writeErr)
		} else {
			fmt.Printf("ℹ No config file found — default config written to %s\n", path)
		}
		return cfg, nil
	}
	if err != nil {
		return AppConfig{}, fmt.Errorf("reading config: %w", err)
	}

	var cfg AppConfig
	if err := json.Unmarshal(data, &cfg); err != nil {
		return AppConfig{}, fmt.Errorf("parsing config: %w", err)
	}
	return cfg, nil
}

// WriteConfig serialises cfg as indented JSON and writes it to path.
func WriteConfig(path string, cfg AppConfig) error {
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}

// TableFields returns only the fields that should be shown as columns.
func (c AppConfig) TableFields() []FieldConfig {
	var out []FieldConfig
	for _, f := range c.Fields {
		if f.ShowInTable {
			out = append(out, f)
		}
	}
	return out
}

// FilterableFields returns only the fields that have an active filter mode.
func (c AppConfig) FilterableFields() []FieldConfig {
	var out []FieldConfig
	for _, f := range c.Fields {
		if f.Filter != FilterNone {
			out = append(out, f)
		}
	}
	return out
}

// FieldByKey looks up a FieldConfig by its JSONKey.
func (c AppConfig) FieldByKey(key string) (FieldConfig, bool) {
	for _, f := range c.Fields {
		if f.JSONKey == key {
			return f, true
		}
	}
	return FieldConfig{}, false
}
