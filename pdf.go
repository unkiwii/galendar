package galendar

import (
	"fmt"
	"os"
	"os/user"
	"path/filepath"
	"runtime"
	"strings"

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

func (r PDFRenderer) createDocument(config Config) (*gofpdf.Fpdf, error) {
	fontDir := getSystemFontDir()
	pdf := gofpdf.New("L", "mm", "A4", fontDir)

	for _, name := range AllFonts {
		font := config.Fonts[name]
		if err := r.registerFont(config, pdf, name, font); err != nil {
			return nil, fmt.Errorf("failed to register font %s: %w", name, err)
		}
	}

	return pdf, nil
}

// RenderMonth renders a single month calendar to PDF
func (r PDFRenderer) RenderMonth(config Config, cal *Calendar) error {
	pdf, err := r.createDocument(config)
	if err != nil {
		return fmt.Errorf("can't create document: %w", err)
	}

	r.renderMonthPage(config, pdf, cal)

	err = pdf.OutputFileAndClose(config.MonthOutputFilePath(cal))
	if err != nil {
		return fmt.Errorf("can't output file: %w", err)
	}

	return nil
}

// RenderYear renders a full year calendar (12 months) to a single PDF
func (r PDFRenderer) RenderYear(config Config, cal *Calendar) error {
	pdf, err := r.createDocument(config)
	if err != nil {
		return fmt.Errorf("can't create document: %w", err)
	}

	// Render each month on a separate page
	for month := 1; month <= 12; month++ {
		cal, err := NewCalendar(cal.Year, month, cal.WeekStart)
		if err != nil {
			return fmt.Errorf("failed to create calendar for month %d: %w", month, err)
		}

		r.renderMonthPage(config, pdf, cal)
	}

	err = pdf.OutputFileAndClose(config.YearOutputFilePath())
	if err != nil {
		return fmt.Errorf("can't output file: %w", err)
	}

	return nil
}

// renderMonthPage renders a single month page
func (r *PDFRenderer) renderMonthPage(config Config, pdf *gofpdf.Fpdf, cal *Calendar) {
	pdf.AddPage()

	pageWidth, pageHeight := pdf.GetPageSize()
	margin := 16.0
	contentWidth := pageWidth - 2*margin
	contentHeight := pageHeight - 2*margin

	// Title (Month Year)
	pdf.SetFont(r.getFontName(config, FontMonths), "B", 24)
	pdf.SetTextColor(0, 0, 0)
	title := fmt.Sprintf("%s %d", config.Language.MonthName(cal.Month), cal.Year)
	titleWidth := pdf.GetStringWidth(title)
	pdf.SetXY((pageWidth/2)-(titleWidth/2), margin)
	pdf.Cell(titleWidth, 15, title)

	// Weekday headers
	pdf.SetFont(r.getFontName(config, FontWeekdays), "B", 22)
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
	}

	// Calendar grid
	pdf.SetFont(r.getFontName(config, FontDays), "", 20)
	pdf.SetTextColor(0, 0, 0)
	gridStartY := headerY + cellHeight
	rows := len(cal.Weeks)
	rowHeight := (contentHeight - (gridStartY - margin)) / float64(rows)

	noteFontSize, noteLineHeight := 0.0, 0.0
	switch rows {
	case 4:
		noteFontSize, noteLineHeight = 18.0, 8.0
	case 5:
		noteFontSize, noteLineHeight = 16.0, 8.0
	case 6:
		noteFontSize, noteLineHeight = 14.0, 6.0
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

			if day.IsCurrentMonth {
				fr, fg, fb, fa := day.FillColor()
				fillStyle := "D"
				if fa != 0 {
					fillStyle = "FD"
					pdf.SetFillColor(fr, fg, fb)
				}

				// Draw number box on current month days only
				pdf.Rect(x, y, cellWidth/3, 12, fillStyle)
			}

			// Draw day number
			dayText := fmt.Sprintf("%d", day.DayNumber)
			numberWidth := pdf.GetStringWidth(dayText)
			if len(dayText) >= 2 {
				numberWidth = 0
			}
			pdf.SetFont(r.getFontName(config, FontDays), "", 20)
			pdf.SetXY(x+1+(numberWidth/2), y-(rowHeight/2)+8)
			pdf.Cell(cellWidth-4, rowHeight-4, dayText)

			if day.Note != "" {
				pdf.SetFont("Times", "I", noteFontSize)
				pdf.SetXY(x+1, y-(rowHeight/2)+10+noteFontSize+(noteLineHeight/4))
				pdf.MultiCell(cellWidth, noteLineHeight, day.Note, "", "L", false)
			}
		}
	}
}

// registerFont registers a font with gofpdf, supporting both font files and built-in fonts
func (r PDFRenderer) registerFont(config Config, pdf *gofpdf.Fpdf, fontKey, fontSpec string) error {
	// It's a file path - try to register it as a TTF font
	ext := strings.ToLower(filepath.Ext(fontSpec))
	if ext == ".ttf" || ext == ".otf" {
		// Use AddUTF8Font to register TTF/OTF fonts
		// The font will be registered with the key we provide
		fontName := r.getFontName(config, fontKey)
		pdf.AddUTF8Font(fontName, "", fontSpec)
		return pdf.Error()
	}
	// If it's not a TTF/OTF, fall through to built-in font mapping

	// Not a file or file doesn't exist - try to use built-in fonts
	// gofpdf has built-in fonts: Courier, Helvetica, Times, Symbol, ZapfDingbats
	// Map common font names to gofpdf built-ins
	builtInFont := r.mapToBuiltInFont(fontSpec)
	if builtInFont != "" {
		// Built-in fonts don't need registration
		return nil
	}

	// If we can't map it, use Helvetica as fallback
	return nil
}

// getFontName returns the font name to use with SetFont
func (r PDFRenderer) getFontName(config Config, fontName string) string {
	fontSpec := config.Fonts[fontName]

	ext := strings.ToLower(filepath.Ext(fontSpec))
	if ext == ".ttf" || ext == ".otf" {
		return fontName
	}

	builtIn := r.mapToBuiltInFont(fontSpec)
	if builtIn != "" {
		return builtIn
	}

	return "Helvetica"
}

// mapToBuiltInFont maps common font names to gofpdf built-in fonts
// gofpdf built-in fonts: Courier, Helvetica, Times, Symbol, ZapfDingbats
func (r PDFRenderer) mapToBuiltInFont(fontName string) string {
	fontLower := strings.ToLower(strings.TrimSpace(fontName))

	// Map common font names to gofpdf built-ins
	fontMap := map[string]string{
		// Helvetica family
		"helvetica":       "Helvetica",
		"arial":           "Helvetica",
		"freesans":        "Helvetica",
		"sans-serif":      "Helvetica",
		"dejavu sans":     "Helvetica",
		"liberation sans": "Helvetica",

		// Times family
		"times":            "Times",
		"times new roman":  "Times",
		"times-roman":      "Times",
		"serif":            "Times",
		"dejavu serif":     "Times",
		"liberation serif": "Times",

		// Courier family
		"courier":         "Courier",
		"courier new":     "Courier",
		"monospace":       "Courier",
		"mono":            "Courier",
		"dejavu mono":     "Courier",
		"liberation mono": "Courier",
	}

	if mapped, ok := fontMap[fontLower]; ok {
		return mapped
	}

	// Default fallback
	return "Helvetica"
}

// getSystemFontDir returns the system font directory based on the operating system
func getSystemFontDir() string {
	switch runtime.GOOS {
	case "windows":
		// Windows font directory
		return "C:\\Windows\\Fonts"
	case "darwin":
		// macOS font directories - try system first, then user
		systemFontDir := "/Library/Fonts"
		if _, err := os.Stat(systemFontDir); err == nil {
			return systemFontDir
		}
		// Fallback to user fonts
		if usr, err := user.Current(); err == nil {
			userFontDir := filepath.Join(usr.HomeDir, "Library", "Fonts")
			if _, err := os.Stat(userFontDir); err == nil {
				return userFontDir
			}
		}
		return systemFontDir
	case "linux":
		// Linux font directories - check common locations
		fontDirs := []string{
			"/usr/share/fonts",
			"/usr/local/share/fonts",
		}

		// Check user fonts
		if usr, err := user.Current(); err == nil {
			userFontDir := filepath.Join(usr.HomeDir, ".fonts")
			if _, err := os.Stat(userFontDir); err == nil {
				return userFontDir
			}
			// Also check .local/share/fonts (more modern location)
			localFontDir := filepath.Join(usr.HomeDir, ".local", "share", "fonts")
			if _, err := os.Stat(localFontDir); err == nil {
				return localFontDir
			}
		}

		// Return first existing system font directory
		for _, dir := range fontDirs {
			if _, err := os.Stat(dir); err == nil {
				return dir
			}
		}

		// Default fallback
		return "/usr/share/fonts"
	default:
		// Unknown OS - return empty string (gofpdf will use default behavior)
		return ""
	}
}
