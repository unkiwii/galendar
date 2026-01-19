package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/unkiwii/galendar/internal/calendar"
	"github.com/unkiwii/galendar/internal/config"
	"github.com/unkiwii/galendar/internal/renderer"
)

func main() {
	var (
		monthFlag      = flag.Int("month", 0, "Month (1-12), 0 means current month")
		yearFlag       = flag.Int("year", 0, "Year, 0 means current year")
		outputFlag     = flag.String("output", "pdf", "Output format: pdf or svg")
		fontMonthFlag  = flag.String("font-month", "", "Font for month name (system font name or path to font file)")
		fontDaysFlag   = flag.String("font-days", "", "Font for day numbers (system font name or path to font file)")
		weekStartFlag  = flag.String("week-start", "sunday", "Week start day: 0-6 (0=Sunday) or day name (sunday, monday, etc.)")
		configFlag     = flag.String("config", "", "Path to JSON configuration file")
		outputPathFlag = flag.String("o", "", "Output file path (directory for SVG year, file for PDF)")
	)

	flag.Parse()

	// Load configuration
	cfg, err := loadConfig(*configFlag, *monthFlag, *yearFlag, *outputFlag, *fontMonthFlag, *fontDaysFlag, *weekStartFlag, *outputPathFlag)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading configuration: %v\n", err)
		os.Exit(1)
	}

	// Determine if we're generating a year or a month
	generateYear := *yearFlag != 0 && *monthFlag == 0

	if generateYear {
		if err := generateYearCalendar(cfg); err != nil {
			fmt.Fprintf(os.Stderr, "Error generating year calendar: %v\n", err)
			os.Exit(1)
		}
	} else {
		if err := generateMonthCalendar(cfg); err != nil {
			fmt.Fprintf(os.Stderr, "Error generating calendar: %v\n", err)
			os.Exit(1)
		}
	}
}

func loadConfig(configPath string, monthFlag, yearFlag int, outputFlag, fontMonthFlag, fontDaysFlag, weekStartFlag, outputPathFlag string) (config.Config, error) {
	var cfg config.Config

	// Load from file if provided
	if configPath != "" {
		fileCfg, err := config.LoadFromFile(configPath)
		if err != nil {
			return config.Config{}, fmt.Errorf("failed to load config file: %w", err)
		}
		cfg = fileCfg
	} else {
		cfg = config.Default()
	}

	// Override with command-line flags (flags take precedence)
	if monthFlag != 0 {
		cfg.Month = monthFlag
	}
	if yearFlag != 0 {
		cfg.Year = yearFlag
	}
	if outputFlag != "" {
		cfg.Output = outputFlag
	}
	if fontMonthFlag != "" {
		cfg.FontMonth = fontMonthFlag
	}
	if fontDaysFlag != "" {
		cfg.FontDays = fontDaysFlag
	}
	if outputPathFlag != "" {
		cfg.OutputPath = outputPathFlag
	}

	// Parse week start
	if weekStartFlag != "" {
		weekStart, err := config.ParseWeekStart(weekStartFlag)
		if err != nil {
			return config.Config{}, fmt.Errorf("invalid week-start: %w", err)
		}
		cfg.WeekStart = weekStart
	}

	return cfg, nil
}

func generateMonthCalendar(cfg config.Config) error {
	cal, err := calendar.NewCalendar(cfg.Year, cfg.Month, cfg.WeekStart)
	if err != nil {
		return fmt.Errorf("failed to create calendar: %w", err)
	}

	outputPath := cfg.OutputPath

	// Generate default output path if not provided
	if outputPath == "" {
		ext := ".pdf"
		if cfg.Output == "svg" {
			ext = ".svg"
		}
		outputPath = fmt.Sprintf("calendar-%04d-%02d%s", cfg.Year, cfg.Month, ext)
	}

	switch cfg.Output {
	case "pdf":
		pdfRenderer := renderer.NewPDFRenderer(cfg)
		return pdfRenderer.RenderMonth(cal, outputPath)
	case "svg":
		svgRenderer := renderer.NewSVGRenderer(cfg)
		return svgRenderer.RenderMonth(cal, outputPath)
	default:
		return fmt.Errorf("unsupported output format: %s (must be 'pdf' or 'svg')", cfg.Output)
	}
}

func generateYearCalendar(cfg config.Config) error {
	outputPath := cfg.OutputPath

	// Generate default output path if not provided
	if outputPath == "" {
		if cfg.Output == "pdf" {
			outputPath = fmt.Sprintf("calendar-%04d.pdf", cfg.Year)
		} else {
			outputPath = fmt.Sprintf("calendar-%04d", cfg.Year)
		}
	}

	switch cfg.Output {
	case "pdf":
		pdfRenderer := renderer.NewPDFRenderer(cfg)
		return pdfRenderer.RenderYear(cfg.Year, cfg.WeekStart, outputPath)
	case "svg":
		svgRenderer := renderer.NewSVGRenderer(cfg)
		// For SVG, outputPath should be a base path (without extension)
		basePath := strings.TrimSuffix(outputPath, filepath.Ext(outputPath))
		return svgRenderer.RenderYear(cfg.Year, cfg.WeekStart, basePath)
	default:
		return fmt.Errorf("unsupported output format: %s (must be 'pdf' or 'svg')", cfg.Output)
	}
}
