package galendar

import (
	"fmt"

	"github.com/jung-kurt/gofpdf"
)

// Renderer handles PDF calendar generation
type PDFRenderer struct{}

func init() {
	RegisterRenderer(PDFRenderer{})
}

func (r PDFRenderer) Name() string {
	return "pdf"
}

// RenderMonth renders a single month calendar to PDF
func (r PDFRenderer) RenderMonth(config Config, cal *Calendar) error {
	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.AddPage()

	monthFont := config.FontMonth
	daysFont := config.FontDays

	// Render the month
	r.renderMonthPage(pdf, cal, monthFont, daysFont)

	return pdf.OutputFileAndClose(config.MonthOutputFilePath(cal))
}

// RenderYear renders a full year calendar (12 months) to a single PDF
func (r PDFRenderer) RenderYear(config Config, cal *Calendar) error {
	pdf := gofpdf.New("L", "mm", "A4", "")

	monthFont := config.FontMonth
	daysFont := config.FontDays

	// Render each month on a separate page
	for month := 1; month <= 12; month++ {
		pdf.AddPage()

		cal, err := NewCalendar(cal.Year, month, cal.WeekStart)
		if err != nil {
			return fmt.Errorf("failed to create calendar for month %d: %w", month, err)
		}

		r.renderMonthPage(pdf, cal, monthFont, daysFont)
	}

	return pdf.OutputFileAndClose(config.YearOutputFilePath())
}

// renderMonthPage renders a single month page
func (r *PDFRenderer) renderMonthPage(pdf *gofpdf.Fpdf, month *Calendar, monthFont, daysFont string) {
	pageWidth, pageHeight := pdf.GetPageSize()
	margin := 20.0
	contentWidth := pageWidth - 2*margin
	contentHeight := pageHeight - 2*margin

	// Title (Month Year)
	pdf.SetFont(monthFont, "B", 24)
	title := fmt.Sprintf("%s %d", month.MonthName, month.Year)
	titleWidth := pdf.GetStringWidth(title)
	pdf.SetXY(margin, margin)
	pdf.Cell(titleWidth, 15, title)

	// Weekday headers
	pdf.SetFont(daysFont, "B", 12)
	weekdayNames := GetWeekdayAbbreviations(month.WeekStart)
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

	for weekIdx, week := range month.Weeks {
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
