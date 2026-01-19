package renderer

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/unkiwii/galendar/internal/calendar"
	"github.com/unkiwii/galendar/internal/config"
)

// SVGRenderer handles SVG calendar generation
type SVGRenderer struct {
	config config.Config
}

// NewSVGRenderer creates a new SVG renderer
func NewSVGRenderer(cfg config.Config) *SVGRenderer {
	return &SVGRenderer{config: cfg}
}

// RenderMonth renders a single month calendar to SVG
func (r *SVGRenderer) RenderMonth(cal *calendar.Calendar, outputPath string) error {
	svg := r.generateSVG(cal)
	return os.WriteFile(outputPath, []byte(svg), 0644)
}

// RenderYear renders a full year calendar, creating 12 separate SVG files
func (r *SVGRenderer) RenderYear(year int, weekStart time.Weekday, basePath string) error {
	// Remove extension if present
	basePath = strings.TrimSuffix(basePath, filepath.Ext(basePath))

	for month := 1; month <= 12; month++ {
		cal, err := calendar.NewCalendar(year, month, weekStart)
		if err != nil {
			return fmt.Errorf("failed to create calendar for month %d: %w", month, err)
		}

		// Generate filename: basePath-YYYY-MM.svg
		outputPath := fmt.Sprintf("%s-%04d-%02d.svg", basePath, year, month)

		if err := r.RenderMonth(cal, outputPath); err != nil {
			return fmt.Errorf("failed to render month %d: %w", month, err)
		}
	}

	return nil
}

// generateSVG generates the SVG content for a calendar
func (r *SVGRenderer) generateSVG(cal *calendar.Calendar) string {
	width := 800
	height := 600
	margin := 40

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf(`<svg width="%d" height="%d" xmlns="http://www.w3.org/2000/svg">`, width, height))
	sb.WriteString("\n")

	// Background
	sb.WriteString(`  <rect width="100%" height="100%" fill="white"/>`)
	sb.WriteString("\n")

	// Title (Month Year)
	monthFont := r.config.FontMonth
	titleY := margin + 30
	sb.WriteString(fmt.Sprintf(`  <text x="%d" y="%d" font-family="%s" font-size="32" font-weight="bold" fill="black">%s %d</text>`,
		width/2, titleY, monthFont, cal.MonthName, cal.Year))
	sb.WriteString("\n")

	// Weekday headers
	daysFont := r.config.FontDays
	cellWidth := (width - 2*margin) / 7
	headerY := titleY + 40

	weekdayNames := calendar.GetWeekdayAbbreviations(cal.WeekStart)
	for i, dayName := range weekdayNames {
		x := margin + i*cellWidth + cellWidth/2
		sb.WriteString(fmt.Sprintf(`  <text x="%d" y="%d" font-family="%s" font-size="14" font-weight="bold" text-anchor="middle" fill="black">%s</text>`,
			x, headerY, daysFont, dayName))
		sb.WriteString("\n")
	}

	// Calendar grid
	gridStartY := headerY + 20
	rowHeight := (height - gridStartY - margin) / 6

	for weekIdx, week := range cal.Weeks {
		for dayIdx, day := range week {

			x := margin + dayIdx*cellWidth
			y := gridStartY + weekIdx*rowHeight

			// Draw cell border
			sb.WriteString(fmt.Sprintf(`  <rect x="%d" y="%d" width="%d" height="%d" fill="white" stroke="#c8c8c8" stroke-width="1"/>`,
				x, y, cellWidth, rowHeight))
			sb.WriteString("\n")
			if !day.IsCurrentMonth {
				// do not draw days outside the current month
				continue
			}

			// Draw day number
			textColor := "black"
			textX := x + 5
			textY := y + 20
			sb.WriteString(fmt.Sprintf(`  <text x="%d" y="%d" font-family="%s" font-size="12" fill="%s">%d</text>`,
				textX, textY, daysFont, textColor, day.DayNumber))
			sb.WriteString("\n")
		}
	}

	sb.WriteString("</svg>")
	return sb.String()
}
