package galendar

import (
	"fmt"
	"log"
	"path/filepath"
	"strings"

	"github.com/adrg/sysfont"
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
func (PDFRenderer) RenderMonth(config Config, cal *Calendar) error {
	pdf, err := createDocument(config)
	if err != nil {
		return fmt.Errorf("can't create document: %w", err)
	}

	err = renderMonthPage(pdf, config, cal)
	if err != nil {
		return fmt.Errorf("failed to render month page %d: %w", cal.Month, err)
	}

	err = pdf.OutputFileAndClose(config.MonthOutputFilePath(cal))
	if err != nil {
		return fmt.Errorf("can't output file: %w", err)
	}

	return nil
}

// RenderYear renders a full year calendar (12 months) to a single PDF
func (PDFRenderer) RenderYear(config Config, cal *Calendar) error {
	pdf, err := createDocument(config)
	if err != nil {
		return fmt.Errorf("can't create document: %w", err)
	}

	// Render each month on a separate page
	for month := 1; month <= 12; month++ {
		cal, err := NewCalendar(cal.Year, month, cal.WeekStart)
		if err != nil {
			return fmt.Errorf("failed to create calendar for month %d: %w", month, err)
		}

		err = renderMonthPage(pdf, config, cal)
		if err != nil {
			return fmt.Errorf("failed to render month page %d: %w", month, err)
		}
	}

	err = pdf.OutputFileAndClose(config.YearOutputFilePath())
	if err != nil {
		return fmt.Errorf("can't output file: %w", err)
	}

	return nil
}

func renderMonthPage(pdf *gofpdf.Fpdf, config Config, cal *Calendar) error {
	pdf.AddPage()

	pageWidth, pageHeight := pdf.GetPageSize()
	margin := 16.0
	contentWidth := pageWidth - 2*margin
	contentHeight := pageHeight - 2*margin

	// Title (Month Year)
	setFont(pdf, FontMonths, 24)
	if err := pdf.Error(); err != nil {
		return fmt.Errorf("can't set font %q: %w", FontMonths, err)
	}
	pdf.SetTextColor(0, 0, 0)
	title := fmt.Sprintf("%s %d", config.Language.MonthName(cal.Month), cal.Year)
	titleWidth := pdf.GetStringWidth(title)
	pdf.SetXY((pageWidth/2)-(titleWidth/2), margin)
	pdf.Cell(titleWidth, 15, title)
	if err := pdf.Error(); err != nil {
		return fmt.Errorf("can't write cell %q: %w", title, err)
	}

	// Weekday headers
	setFont(pdf, FontWeekdays, 22)
	if err := pdf.Error(); err != nil {
		return fmt.Errorf("can't set font %q: %w", FontWeekdays, err)
	}
	pdf.SetTextColor(0, 0, 0)
	weekdayNames := config.Language.WeekdayAbbreviations(cal.WeekStart)
	cellWidth := contentWidth / 7
	cellHeight := 10.0
	headerY := (margin * 2.2)

	for i, dayName := range weekdayNames {
		dayWidth := pdf.GetStringWidth(dayName)
		x := (margin + float64(i)*cellWidth) + (cellWidth / 2) - (dayWidth / 2)

		pdf.SetTextColor(0, 0, 0)
		pdf.SetXY(x, headerY)
		pdf.Cell(cellWidth, cellHeight, dayName)
		if err := pdf.Error(); err != nil {
			return fmt.Errorf("can't write cell %q: %w", dayName, err)
		}
	}

	// Calendar grid
	gridStartY := headerY + cellHeight
	rows := len(cal.Weeks)
	rowHeight := (contentHeight - (gridStartY - margin)) / float64(rows)

	noteFontSize, noteLineHeight := config.FontSizes[FontNotes], 0.0
	switch rows {
	case 4:
		noteFontSize, noteLineHeight = noteFontSize, (noteFontSize/2)-1
	case 5:
		noteFontSize, noteLineHeight = noteFontSize-2, (noteFontSize/2)-1
	case 6:
		noteFontSize, noteLineHeight = noteFontSize-4, (noteFontSize/2)-3
	}

	for weekIdx, week := range cal.Weeks {
		for dayIdx, day := range week {
			x := margin + float64(dayIdx)*cellWidth
			y := gridStartY + float64(weekIdx)*rowHeight

			// Draw cell border
			pdf.SetDrawColor(150, 150, 150)
			pdf.Rect(x, y, cellWidth, rowHeight, "D")

			tr, tg, tb, ta := day.TextColor()
			if ta == 0 && !config.ShowExtraDays {
				continue
			}
			pdf.SetTextColor(tr, tg, tb)

			dayBoxHeight := 12.0
			dayBoxBottom := y + dayBoxHeight
			if day.IsCurrentMonth {
				fr, fg, fb, fa := day.FillColor()
				fillStyle := "D"
				if fa != 0 {
					fillStyle = "FD"
					pdf.SetFillColor(fr, fg, fb)
				}

				// Draw number box on current month days only
				pdf.Rect(x, y, cellWidth/3, dayBoxHeight, fillStyle)
			}

			// Draw day number
			dayText := fmt.Sprintf("%d", day.DayNumber)
			numberWidth := pdf.GetStringWidth(dayText)
			if len(dayText) >= 2 {
				numberWidth = 0
			}
			setFont(pdf, FontDays, 20)
			if err := pdf.Error(); err != nil {
				return fmt.Errorf("can't set font %q: %w", FontWeekdays, err)
			}
			pdf.SetXY(x+1+(numberWidth/2), y-(rowHeight/2)+8)
			pdf.Cell(cellWidth-4, rowHeight-4, dayText)
			if err := pdf.Error(); err != nil {
				return fmt.Errorf("can't write cell %q: %w", dayText, err)
			}

			if day.Note != nil {
				noteSize := noteFontSize
				noteHeight := noteLineHeight
				if day.Note.Size != 0 {
					noteSize = day.Note.Size
					noteHeight = (noteSize / 2) - 1
				}
				if day.Note.Font != "" {
					if err := registerFont(pdf, day.Name(), day.Note.Font); err != nil {
						return fmt.Errorf("failed to register font %s: %w", day.Name(), err)
					}
					setFont(pdf, day.Name(), noteSize)
				} else {
					setFont(pdf, FontNotes, noteSize)
				}
				if err := pdf.Error(); err != nil {
					return fmt.Errorf("can't set font %q: %w", FontNotes, err)
				}
				pdf.SetXY(x+1, dayBoxBottom+2)
				pdf.MultiCell(cellWidth, noteHeight, day.Note.Text, "", "L", false)
				if err := pdf.Error(); err != nil {
					return fmt.Errorf("can't write multi cell %q: %w", day.Note, err)
				}
			}
		}
	}

	return pdf.Error()
}

func createDocument(config Config) (*gofpdf.Fpdf, error) {
	pdf := gofpdf.New("L", "mm", "A4", "")

	for _, name := range AllFonts {
		font := config.Fonts[name]
		if err := registerFont(pdf, name, font); err != nil {
			return nil, fmt.Errorf("failed to register font %s: %w", name, err)
		}
	}

	return pdf, nil
}

func registerFont(pdf *gofpdf.Fpdf, internalFontName, fontName string) error {
	ext := strings.ToLower(filepath.Ext(fontName))
	if ext == ".ttf" || ext == ".otf" {
		return registerFontFile(pdf, internalFontName, fontName)
	}

	fontsFinder := sysfont.NewFinder(nil)
	font := fontsFinder.Match(fontName)
	if font == nil {
		return fmt.Errorf("font %s (%q) not found", fontName, internalFontName)
	}

	return registerFontFile(pdf, internalFontName, font.Filename)
}

var registeredFontsStyle map[string]string

func registerFontFile(pdf *gofpdf.Fpdf, fontName, filename string) error {
	log.Printf("registerFontFile(%s, %s)", fontName, filename)
	style := ""
	if strings.Contains(filename, "Italic") || strings.Contains(filename, "Ita") {
		style += "I"
	}
	if strings.Contains(filename, "Bold") || strings.Contains(filename, "Bd") {
		style += "B"
	}

	if registeredFontsStyle == nil {
		registeredFontsStyle = map[string]string{}
	}
	registeredFontsStyle[fontName] = style

	pdf.SetFontLocation(filepath.Dir(filename))
	pdf.AddUTF8Font(fontName, style, filepath.Base(filename))
	return pdf.Error()
}

// setFont tries to set a font with 3 different styles: Regular, Italic and
// Bold, it sets the first that doesn't errors out, if all 3 errors
func setFont(pdf *gofpdf.Fpdf, name string, size float64) error {
	if pdf.Error() != nil {
		return pdf.Error()
	}

	style, ok := registeredFontsStyle[name]
	if !ok {
		style = ""
	}
	pdf.SetFont(name, style, size)

	return pdf.Error()
}
