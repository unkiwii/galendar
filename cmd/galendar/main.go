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
		weekStartFlag  = flag.String("week-start", "sunday", "Week start day: sunday or monday")
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
		year := cfg.GetYear()
		if err := generateYearCalendar(cfg, year); err != nil {
			fmt.Fprintf(os.Stderr, "Error generating year calendar: %v\n", err)
			os.Exit(1)
		}
	} else {
		month := cfg.GetMonth()
		year := cfg.GetYear()
		if err := generateMonthCalendar(cfg, year, month); err != nil {
			fmt.Fprintf(os.Stderr, "Error generating calendar: %v\n", err)
			os.Exit(1)
		}
	}
}

func loadConfig(configPath string, monthFlag, yearFlag int, outputFlag, fontMonthFlag, fontDaysFlag, weekStartFlag, outputPathFlag string) (*config.Config, error) {
	var cfg *config.Config

	// Load from file if provided
	if configPath != "" {
		fileCfg, err := config.LoadFromFile(configPath)
		if err != nil {
			return nil, fmt.Errorf("failed to load config file: %w", err)
		}
		cfg = fileCfg
	} else {
		cfg = &config.Config{}
	}

	// Override with command-line flags (flags take precedence)
	if monthFlag != 0 {
		cfg.Month = &monthFlag
	}
	if yearFlag != 0 {
		cfg.Year = &yearFlag
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
		switch strings.ToLower(weekStartFlag) {
		case "sunday", "sun":
			cfg.WeekStart = config.Sunday
		case "monday", "mon":
			cfg.WeekStart = config.Monday
		default:
			return nil, fmt.Errorf("invalid week-start: %s (must be 'sunday' or 'monday')", weekStartFlag)
		}
	}

	return cfg, nil
}

func generateMonthCalendar(cfg *config.Config, year, month int) error {
	cal, err := calendar.NewCalendar(year, month, cfg.GetWeekStart())
	if err != nil {
		return fmt.Errorf("failed to create calendar: %w", err)
	}

	outputFormat := cfg.GetOutputFormat()
	outputPath := cfg.OutputPath

	// Generate default output path if not provided
	if outputPath == "" {
		ext := ".pdf"
		if outputFormat == "svg" {
			ext = ".svg"
		}
		outputPath = fmt.Sprintf("calendar-%04d-%02d%s", year, month, ext)
	}

	switch outputFormat {
	case "pdf":
		pdfRenderer := renderer.NewPDFRenderer(cfg)
		return pdfRenderer.RenderMonth(cal, outputPath)
	case "svg":
		svgRenderer := renderer.NewSVGRenderer(cfg)
		return svgRenderer.RenderMonth(cal, outputPath)
	default:
		return fmt.Errorf("unsupported output format: %s (must be 'pdf' or 'svg')", outputFormat)
	}
}

func generateYearCalendar(cfg *config.Config, year int) error {
	outputFormat := cfg.GetOutputFormat()
	outputPath := cfg.OutputPath

	// Generate default output path if not provided
	if outputPath == "" {
		if outputFormat == "pdf" {
			outputPath = fmt.Sprintf("calendar-%04d.pdf", year)
		} else {
			outputPath = fmt.Sprintf("calendar-%04d", year)
		}
	}

	switch outputFormat {
	case "pdf":
		pdfRenderer := renderer.NewPDFRenderer(cfg)
		return pdfRenderer.RenderYear(year, cfg.GetWeekStart(), outputPath)
	case "svg":
		svgRenderer := renderer.NewSVGRenderer(cfg)
		// For SVG, outputPath should be a base path (without extension)
		basePath := strings.TrimSuffix(outputPath, filepath.Ext(outputPath))
		return svgRenderer.RenderYear(year, cfg.GetWeekStart(), basePath)
	default:
		return fmt.Errorf("unsupported output format: %s (must be 'pdf' or 'svg')", outputFormat)
	}
}
