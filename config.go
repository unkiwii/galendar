package galendar

import (
	"fmt"
	"os"
	"path"
	"strings"
	"time"

	"github.com/spf13/viper"
)

const (
	DefaultFont      = "courier"
	DefaultWeekStart = time.Sunday
)

// Config holds the application configuration with all values already resolved
type Config struct {
	Month         int          // 1-12, 0 means current month
	Year          int          // 0 means current year
	WeekStart     time.Weekday // 0-6, representing Sunday through Saturday
	FontMonth     string       // Font name or path for month
	FontDays      string       // Font name or path for days
	FontNotes     string       // Font name or path for notes
	Renderer      Renderer     // "pdf" or "svg", default "pdf"
	OutputDir     string       // Output directory name
	ShowExtraDays bool         // show days outside current month (defaults to false)
	Language      Language     // language to use on the output (defaults to Spanish)
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

	return 0, fmt.Errorf("invalid week start: %q (must be 0-6 or a day name)", s)
}

func NewConfig(v *viper.Viper) (Config, error) {
	weekStart, err := ParseWeekStart(viper.GetString("week-start"))
	if err != nil {
		return Config{}, fmt.Errorf("invalid week start: %w", err)
	}

	renderer, err := RendererByName(viper.GetString("renderer"))
	if err != nil {
		return Config{}, fmt.Errorf("invalid renderer: %w", err)
	}

	outputDir := viper.GetString("output-dir")
	if info, err := os.Stat(outputDir); err != nil {
		return Config{}, fmt.Errorf("invalid output dir: %w", err)
	} else if !info.IsDir() {
		return Config{}, fmt.Errorf("invalid output dir: %q is not a directory", outputDir)
	}

	language := Language(viper.GetString("language"))
	if !IsValidLanguage(language) {
		return Config{}, fmt.Errorf("invalid language: %q", language)
	}

	return Config{
		Month:         viper.GetInt("month"),
		Year:          viper.GetInt("year"),
		WeekStart:     weekStart,
		FontMonth:     viper.GetString("font-month"),
		FontDays:      viper.GetString("font-days"),
		FontNotes:     viper.GetString("font-notes"),
		Renderer:      renderer,
		OutputDir:     outputDir,
		ShowExtraDays: viper.GetBool("show-extra-days"),
		Language:      language,
	}, nil
}

func (cfg Config) YearOutputFilePath() string {
	filename := fmt.Sprintf("%s-%04d.%s", cfg.Language.Read("calendar"), cfg.Year, cfg.Renderer.Name())
	return path.Join(cfg.OutputDir, filename)
}

func (cfg Config) MonthOutputFilePath(cal *Calendar) string {
	filename := fmt.Sprintf("%s-%04d-%02d.%s", cfg.Language.Read("calendar"), cfg.Year, cal.Month, cfg.Renderer.Name())
	return path.Join(cfg.OutputDir, filename)
}
