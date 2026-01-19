package config

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"
)

const DefaultFont = "Arial"

type OutputType string

// TODO: let renderers "register" their output type
const (
	OutputTypePDF = "pdf"
	OutputTypeSVG = "svg"
)

// Config holds the application configuration with all values already resolved
type Config struct {
	Month      int          // 1-12, 0 means current month
	Year       int          // 0 means current year
	WeekStart  time.Weekday // 0-6, representing Sunday through Saturday
	FontMonth  string       // Font name or path for month
	FontDays   string       // Font name or path for days
	OutputType OutputType   // "pdf" or "svg", default "pdf"
	OutputPath string       // Output directory path
}

var weekdayStringToWeekday = map[string]time.Weekday{
	"sunday":    time.Sunday,
	"sun":       time.Sunday,
	"monday":    time.Monday,
	"mon":       time.Monday,
	"tuesday":   time.Tuesday,
	"tue":       time.Tuesday,
	"wednesday": time.Wednesday,
	"wed":       time.Wednesday,
	"thursday":  time.Thursday,
	"thu":       time.Thursday,
	"friday":    time.Friday,
	"fri":       time.Friday,
	"saturday":  time.Saturday,
	"sat":       time.Saturday,
}

var outputTypeStringToOutputType = map[string]OutputType{
	"pdf": OutputTypePDF,
	"svg": OutputTypeSVG,
}

// ParseWeekStart parses a week start string into a WeekStart value
// Accepts: day names (sunday, monday, etc.), abbreviations (sun, mon, etc.), or numbers (0-6)
func ParseWeekStart(s string) (time.Weekday, error) {
	s = strings.ToLower(strings.TrimSpace(s))

	// Try numeric value first
	if len(s) == 1 && s >= "0" && s <= "6" {
		return time.Weekday(int(s[0] - '0')), nil
	}

	if val, ok := weekdayStringToWeekday[s]; ok {
		return val, nil
	}

	return 0, fmt.Errorf("invalid week start: %s (must be 0-6 or a day name)", s)
}

func ParseOutputType(s string) (OutputType, error) {
	s = strings.ToLower(strings.TrimSpace(s))

	if val, ok := outputTypeStringToOutputType[s]; ok {
		return val, nil
	}

	return "", fmt.Errorf("invalid output type: %s (must be pdf or svg)", s)
}

// LoadFromFile loads configuration from a JSON file and returns a fully resolved Config
func LoadFromFile(path string) (Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return Config{}, fmt.Errorf("failed to read config file: %w", err)
	}

	var fileConfig struct {
		Month      *int    `json:"month,omitempty"`
		Year       *int    `json:"year,omitempty"`
		FontMonth  string  `json:"font_month,omitempty"`
		FontDays   string  `json:"font_days,omitempty"`
		WeekStart  *string `json:"week_start,omitempty"`
		OutputType string  `json:"output_type,omitempty"`
		OutputPath string  `json:"output_path,omitempty"`
	}

	if err := json.Unmarshal(data, &fileConfig); err != nil {
		return Config{}, fmt.Errorf("failed to parse config file: %w", err)
	}

	cfg := Default()

	if fileConfig.Month != nil {
		cfg.Month = *fileConfig.Month
	}
	if fileConfig.Year != nil {
		cfg.Year = *fileConfig.Year
	}
	if isValidFontFile(fileConfig.FontMonth) {
		cfg.FontMonth = fileConfig.FontMonth
	}
	if isValidFontFile(fileConfig.FontDays) {
		cfg.FontDays = fileConfig.FontDays
	}
	if fileConfig.OutputPath != "" {
		cfg.OutputPath = fileConfig.OutputPath
	}

	if fileConfig.OutputType != "" {
		outputType, err := ParseOutputType(fileConfig.OutputType)
		if err != nil {
			return Config{}, fmt.Errorf("invalid output_type in config: %w", err)
		}
		cfg.OutputType = outputType
	}

	if fileConfig.WeekStart != nil {
		weekStart, err := ParseWeekStart(*fileConfig.WeekStart)
		if err != nil {
			return Config{}, fmt.Errorf("invalid week_start in config: %w", err)
		}
		cfg.WeekStart = weekStart
	}

	return cfg, nil
}

func isValidFontFile(filename string) bool {
	if filename == "" {
		return false
	}

	if _, err := os.Stat(filename); err == nil {
		return false
	}

	return true
}

// Default returns a Config with default values (current month/year, PDF output, Sunday week start)
func Default() Config {
	now := time.Now()

	return Config{
		Month:      int(now.Month()),
		Year:       now.Year(),
		WeekStart:  time.Sunday,
		FontMonth:  DefaultFont,
		FontDays:   DefaultFont,
		OutputType: OutputTypePDF,
		OutputPath: "",
	}
}
