package config

import (
	"encoding/json"
	"fmt"
	"os"
	"time"
)

// WeekStart represents the starting day of the week
type WeekStart int

const (
	Sunday WeekStart = iota
	Monday
)

// Config holds the application configuration
type Config struct {
	Month      *int      `json:"month,omitempty"`       // 1-12, nil means current month
	Year       *int      `json:"year,omitempty"`        // nil means current year
	Output     string    `json:"output,omitempty"`      // "pdf" or "svg", default "pdf"
	FontMonth  string    `json:"font_month,omitempty"`  // Font name or path for month
	FontDays   string    `json:"font_days,omitempty"`   // Font name or path for days
	WeekStart  WeekStart `json:"week_start,omitempty"`  // 0 = Sunday, 1 = Monday
	OutputPath string    `json:"output_path,omitempty"` // Output directory path
}

// LoadFromFile loads configuration from a JSON file
func LoadFromFile(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	return &cfg, nil
}

// GetMonth returns the month to use (1-12)
func (c *Config) GetMonth() int {
	if c.Month != nil {
		return *c.Month
	}
	return int(time.Now().Month())
}

// GetYear returns the year to use
func (c *Config) GetYear() int {
	if c.Year != nil {
		return *c.Year
	}
	return time.Now().Year()
}

// GetOutputFormat returns the output format, defaulting to "pdf"
func (c *Config) GetOutputFormat() string {
	if c.Output == "" {
		return "pdf"
	}
	return c.Output
}

// GetWeekStart returns the week start day
func (c *Config) GetWeekStart() WeekStart {
	return c.WeekStart
}
