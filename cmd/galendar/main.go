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
	now := time.Now()

	pflag.Int("month", 0, "Month: 1-12 to render the month, 0 (or missing) to render the whole year")
	pflag.Int("year", now.Year(), "Year")
	pflag.String("renderer", galendar.DefaultRenderer().Name(), "Output format: pdf or svg")
	pflag.String("font-month", galendar.DefaultFont, "Font for month name (system font name or path to font file)")
	pflag.String("font-days", galendar.DefaultFont, "Font for day numbers (system font name or path to font file)")
	pflag.String("week-start", galendar.DefaultWeekStart.String(), "Week start day: 0-6 (0=Sunday) or day name (sunday, monday, etc.)")
	pflag.String("config", "", "Path to JSON configuration file")
	pflag.String("output-dir", "", "Output directory, defaults to current directory")
	pflag.Parse()

	currentDir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("can't read working directory: %w", err)
	}

	viper.SetDefault("month", 0)
	viper.SetDefault("year", now.Year())
	viper.SetDefault("renderer", "pdf")
	viper.SetDefault("font-month", "Arial")
	viper.SetDefault("font-days", "Arial")
	viper.SetDefault("week-start", "sunday")
	viper.SetDefault("output-dir", currentDir)

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
