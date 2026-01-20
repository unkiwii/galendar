package main

import (
	"fmt"
	"os"
	"time"

	"github.com/spf13/pflag"
	"github.com/spf13/viper"

	"github.com/unkiwii/galendar"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}
}

func run() error {
	defaultOutputDir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("can't read working directory: %w", err)
	}

	defaultMonth := 0
	defaultYear := time.Now().Year()
	defaultRenderer := galendar.DefaultRenderer().Name()
	defaultFont := galendar.DefaultFont
	defaultWeekStart := galendar.DefaultWeekStart.String()

	pflag.IntP("month", "m", defaultMonth, "Month: 1-12 to render the month, 0 (or missing) to render the whole year")
	pflag.IntP("year", "y", defaultYear, "Year")
	pflag.String("renderer", defaultRenderer, "Output format: pdf or svg")
	pflag.String("font-month", defaultFont, "Font for month name (system font name or path to font file)")
	pflag.String("font-days", defaultFont, "Font for day numbers (system font name or path to font file)")
	pflag.String("font-notes", defaultFont, "Font for notes (system font name or path to font file)")
	pflag.String("week-start", defaultWeekStart, "Week start day: 0-6 (0=Sunday) or day name (sunday, monday, etc.)")
	pflag.String("config", "", "Path to JSON configuration file")
	pflag.StringP("output-dir", "o", "", "Output directory, defaults to current directory")
	pflag.Parse()

	viper.SetDefault("month", defaultMonth)
	viper.SetDefault("year", defaultYear)
	viper.SetDefault("renderer", defaultRenderer)
	viper.SetDefault("font-month", defaultFont)
	viper.SetDefault("font-days", defaultFont)
	viper.SetDefault("week-start", defaultWeekStart)
	viper.SetDefault("output-dir", defaultOutputDir)

	viper.SetEnvPrefix("galendar")
	viper.AutomaticEnv()

	viper.BindPFlags(pflag.CommandLine)

	configFile := pflag.Lookup("config").Value.String()
	if configFile != "" {
		viper.SetConfigFile(configFile)
		err := viper.ReadInConfig()
		if err != nil {
			return fmt.Errorf("invalid config file: %w", err)
		}
	}

	cfg, err := galendar.NewConfig(viper.GetViper())
	if err != nil {
		return fmt.Errorf("invalid configuration: %w", err)
	}

	return writeCalendar(cfg)
}

func writeCalendar(cfg galendar.Config) error {
	month := cfg.Month

	renderFunc := cfg.Renderer.RenderMonth
	if month == 0 {
		month = 1
		renderFunc = cfg.Renderer.RenderYear
	}

	cal, err := galendar.NewCalendar(cfg.Year, month, cfg.WeekStart)
	if err != nil {
		return fmt.Errorf("invalid calendar: %w", err)
	}

	err = renderFunc(cfg, cal)
	if err != nil {
		return fmt.Errorf("can't generate year calendar: %w", err)
	}

	return nil
}
