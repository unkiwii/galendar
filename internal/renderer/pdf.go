package renderer

import (
	"fmt"
	"os"

	"github.com/jung-kurt/gofpdf"
	"github.com/unkiwii/galendar/internal/calendar"
	"github.com/unkiwii/galendar/internal/config"
)

// PDFRenderer handles PDF calendar generation
type PDFRenderer struct {
	config *config.Config
}

// NewPDFRenderer creates a new PDF renderer
func NewPDFRenderer(cfg *config.Config) *PDFRenderer {
	return &PDFRenderer{config: cfg}
}

// RenderMonth renders a single month calendar to PDF
func (r *PDFRenderer) RenderMonth(cal *calendar.Calendar, outputPath string) error {
	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.AddPage()
	
	// Set up fonts
	monthFont := r.getFont(r.config.FontMonth, "Arial")
	daysFont := r.getFont(r.config.FontDays, "Arial")
	
	// Render the month
	r.renderMonthPage(pdf, cal, monthFont, daysFont)
	
	return pdf.OutputFileAndClose(outputPath)
}

// RenderYear renders a full year calendar (12 months) to a single PDF
func (r *PDFRenderer) RenderYear(year int, weekStart config.WeekStart, outputPath string) error {
	pdf := gofpdf.New("P", "mm", "A4", "")
	
	monthFont := r.getFont(r.config.FontMonth, "Arial")
	daysFont := r.getFont(r.config.FontDays, "Arial")
	
	// Render each month on a separate page
	for month := 1; month <= 12; month++ {
		if month > 1 {
			pdf.AddPage()
		}
		
		cal, err := calendar.NewCalendar(year, month, weekStart)
		if err != nil {
			return fmt.Errorf("failed to create calendar for month %d: %w", month, err)
		}
		
		r.renderMonthPage(pdf, cal, monthFont, daysFont)
	}
	
	return pdf.OutputFileAndClose(outputPath)
}

// renderMonthPage renders a single month page
func (r *PDFRenderer) renderMonthPage(pdf *gofpdf.Fpdf, cal *calendar.Calendar, monthFont, daysFont string) {
	pageWidth, pageHeight := pdf.GetPageSize()
	margin := 20.0
	contentWidth := pageWidth - 2*margin
	contentHeight := pageHeight - 2*margin
	
	// Title (Month Year)
	pdf.SetFont(monthFont, "B", 24)
	title := fmt.Sprintf("%s %d", cal.MonthName, cal.Year)
	titleWidth := pdf.GetStringWidth(title)
	pdf.SetXY(margin, margin)
	pdf.Cell(titleWidth, 15, title)
	
	// Weekday headers
	pdf.SetFont(daysFont, "B", 12)
	weekdayNames := calendar.GetWeekdayAbbreviations(cal.WeekStart)
	cellWidth := contentWidth / 7
	cellHeight := 10.0
	headerY := margin + 25
	
	for i, dayName := range weekdayNames {
		x := margin + float64(i)*cellWidth
		pdf.SetXY(x, headerY)
		pdf.Cell(cellWidth, cellHeight, dayName)
	}
	
	// Calendar grid
	pdf.SetFont(daysFont, "", 10)
	gridStartY := headerY + cellHeight + 5
	rowHeight := (contentHeight - (gridStartY - margin)) / 6
	
	for weekIdx, week := range cal.Weeks {
		for dayIdx, day := range week {
			x := margin + float64(dayIdx)*cellWidth
			y := gridStartY + float64(weekIdx)*rowHeight
			
			// Draw cell border
			pdf.SetDrawColor(200, 200, 200)
			pdf.Rect(x, y, cellWidth, rowHeight, "D")
			
			// Draw day number
			if day.IsCurrentMonth {
				pdf.SetTextColor(0, 0, 0)
			} else {
				pdf.SetTextColor(150, 150, 150)
			}
			
			pdf.SetXY(x+2, y+2)
			pdf.Cell(cellWidth-4, rowHeight-4, fmt.Sprintf("%d", day.DayNumber))
		}
	}
}

// getFont returns the font name to use, handling both system fonts and custom font paths
func (r *PDFRenderer) getFont(fontSpec, defaultFont string) string {
	if fontSpec == "" {
		return defaultFont
	}
	
	// Check if it's a file path
	if _, err := os.Stat(fontSpec); err == nil {
		// It's a file path - we'll need to register it
		// For now, return default and log that custom fonts need font registration
		// This is a limitation we'll note - gofpdf requires font registration
		return defaultFont
	}
	
	// Assume it's a system font name
	return fontSpec
}
