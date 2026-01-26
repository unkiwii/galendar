package galendar

import (
	"encoding/xml"
	"fmt"
	"io"
	"log"
	"os"
	"strings"
)

// SVGRenderer handles SVG calendar generation
type SVGRenderer struct{}

func init() {
	RegisterRenderer(SVGRenderer{})
}

func (r SVGRenderer) Name() string {
	return "svg"
}

// RenderMonth renders a single month calendar to SVG
func (r SVGRenderer) RenderMonth(config Config, cal Calendar) error {
	svg := r.generateSVG(config, cal)
	return os.WriteFile(config.MonthOutputFilePath(cal), []byte(svg), 0644)
}

// RenderYear renders a full year calendar, creating 12 separate SVG files
func (r SVGRenderer) RenderYear(config Config, cal Calendar) error {
	for month := 1; month <= 12; month++ {
		cal, err := cal.CloneAt(month)
		if err != nil {
			return fmt.Errorf("can't clone calendar at month %d: %w", month, err)
		}

		if err := r.RenderMonth(config, cal); err != nil {
			return fmt.Errorf("failed to render month %d: %w", month, err)
		}
	}

	return nil
}

// generateSVG generates the SVG content for a calendar
// TODO: change return type to []byte
func (r SVGRenderer) generateSVG(config Config, cal Calendar) string {
	width := 800
	height := 600
	margin := 40

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf(`<svg width="%d" height="%d" xmlns="http://www.w3.org/2000/svg" xmlns:xlink="http://www.w3.org/1999/xlink">`, width, height))
	sb.WriteString("\n")

	// Background
	sb.WriteString(`  <rect width="100%" height="100%" fill="white"/>`)
	sb.WriteString("\n")

	// Collect unique SVG icons from special days
	iconMap := r.collectSVGIcons(cal)

	// Write defs section with all icons
	if len(iconMap) > 0 {
		r.writeDefsSection(&sb, iconMap)
	}

	// Title (Month Year)
	monthFont := config.Fonts[FontMonths]
	titleY := margin + 30
	sb.WriteString(fmt.Sprintf(`  <text x="%s" y="%d" text-anchor="middle" font-family="%s" font-size="24" font-weight="" fill="black">%s %d</text>`,
		"50%", titleY, monthFont, config.Language.MonthName(cal.Month), cal.Year))
	sb.WriteString("\n")

	// Weekday headers
	daysFont := config.Fonts[FontDays]
	cellWidth := (width - 2*margin) / 7
	headerY := titleY + 40

	weekdayNames := config.Language.WeekdayAbbreviations(cal.WeekStart)
	for i, dayName := range weekdayNames {
		x := margin + i*cellWidth + cellWidth/2
		sb.WriteString(fmt.Sprintf(`  <text x="%d" y="%d" font-family="%s" font-size="22" font-weight="" text-anchor="middle" fill="black">%s</text>`,
			x, headerY, daysFont, dayName))
		sb.WriteString("\n")
	}

	// Calendar grid
	gridStartY := headerY + 20
	rows := len(cal.Weeks)
	rowHeight := float64(height-gridStartY-margin) / float64(rows)

	// Calculate note font size based on number of rows (matching PDF logic)
	noteFontSize := float64(config.FontSizes[FontNotes])
	switch rows {
	case 5:
		noteFontSize = noteFontSize - 2
	case 6:
		noteFontSize = noteFontSize - 4
	}

	for weekIdx, week := range cal.Weeks {
		for dayIdx, day := range week {
			x := margin + dayIdx*cellWidth
			y := gridStartY + weekIdx*int(rowHeight)

			// Draw cell border
			sb.WriteString(fmt.Sprintf(`  <rect x="%d" y="%d" width="%d" height="%.0f" fill="white" stroke="#969696" stroke-width="1"/>`,
				x, y, cellWidth, rowHeight))
			sb.WriteString("\n")

			// Get text and fill colors
			tr, tg, tb, ta := day.TextColor()
			if ta == 0 && !config.ShowExtraDays {
				continue
			}

			// Draw day box rectangle for current month days (matching PDF)
			dayBoxHeight := 36.0
			dayBoxBottom := float64(y) + dayBoxHeight
			if day.IsCurrentMonth {
				fr, fg, fb, fa := day.FillColor()
				fill := "white"
				// Draw filled rectangle for holidays (FD mode in PDF)
				if fa != 0 {
					fill = fmt.Sprintf("rgb(%d,%d,%d)", fr, fg, fb)
				}

				sb.WriteString(fmt.Sprintf(`  <rect x="%d" y="%d" width="%.0f" height="%.0f" fill="%s" stroke="#969696"/>`,
					x, y, float64(cellWidth)/3, dayBoxHeight, fill))
				sb.WriteString("\n")
			}

			// Draw day number (matching PDF positioning)
			dayText := fmt.Sprintf("%d", day.DayNumber)
			textColor := fmt.Sprintf("rgb(%d,%d,%d)", tr, tg, tb)
			// Position similar to PDF: x+1+(numberWidth/2), y-(rowHeight/2)+8
			// For SVG, we approximate this positioning
			numberWidth := 14
			if len(dayText) >= 2 {
				numberWidth = 0
			}
			textX := x + 4 + (numberWidth / 2)
			textY := y + (int(dayBoxHeight) / 2) + 8
			sb.WriteString(fmt.Sprintf(`  <text x="%d" y="%d" font-family="%s" font-size="20" fill="%s">%s</text>`,
				textX, textY, daysFont, textColor, dayText))
			sb.WriteString("\n")

			// Render special day icon if present
			if day.special != nil && day.special.Icon != "" {
				if iconID, ok := iconMap[day.special.Icon]; ok {
					iconSize := cellWidth / 3
					iconX := x + cellWidth - iconSize - 5
					iconY := y + 5
					// Use <use> with symbol - width and height will scale the symbol
					// Use xlink:href for better compatibility with older SVG viewers
					sb.WriteString(fmt.Sprintf(`  <use xlink:href="#%s" x="%d" y="%d" width="%d" height="%d"/>`,
						iconID, iconX, iconY, iconSize, iconSize))
					sb.WriteString("\n")
				}
			}

			// Render special day note/text if present (matching PDF logic)
			if note := day.Note(); note != nil {
				noteSize := noteFontSize
				noteLineHeight := noteSize
				if note.Size != 0 {
					noteSize = note.Size
					noteLineHeight = (noteSize / 2) - 1
				}
				noteFont := config.Fonts[FontNotes]
				if note.Font != "" {
					noteFont = note.Font
				}
				noteX := x + 5
				noteY := int(dayBoxBottom) + 24
				availableWidth := float64(cellWidth - 5) // Leave padding on both sides

				// Break text into lines that fit within the cell width
				lines := r.wrapText(note.Text, noteSize, availableWidth)

				// Render wrapped text using tspan elements
				sb.WriteString(fmt.Sprintf(`  <text x="%d" y="%d" font-family="%s" font-size="%.1f" fill="black">`,
					noteX, noteY, noteFont, noteSize))
				for i, line := range lines {
					if i == 0 {
						// First line uses the base text element
						sb.WriteString(escapeXML(line))
					} else {
						// Subsequent lines use tspan with dy for line spacing
						sb.WriteString(fmt.Sprintf(`<tspan x="%d" dy="%.1f">%s</tspan>`,
							noteX, noteLineHeight, escapeXML(line)))
					}
				}
				sb.WriteString("</text>\n")
			}
		}
	}

	sb.WriteString("</svg>")
	return sb.String()
}

// collectSVGIcons collects all unique SVG icon files from the calendar's special days
func (r SVGRenderer) collectSVGIcons(cal Calendar) map[string]string {
	iconMap := make(map[string]string)
	iconCounter := 0

	for _, week := range cal.Weeks {
		for _, day := range week {
			if day.special != nil && day.special.Icon != "" {
				iconPath := day.special.Icon
				// Only add if not already in map
				if _, exists := iconMap[iconPath]; !exists {
					iconID := fmt.Sprintf("icon-%d", iconCounter)
					iconMap[iconPath] = iconID
					iconCounter++
				}
			}
		}
	}

	log.Println("icons found:")
	for k, v := range iconMap {
		log.Printf("  %s : %s", k, v)
	}

	return iconMap
}

// writeDefsSection writes the <defs> section with all SVG icons
func (r SVGRenderer) writeDefsSection(sb *strings.Builder, iconMap map[string]string) {
	sb.WriteString("  <defs>\n")

	for iconPath, iconID := range iconMap {
		innerContent, viewBox, err := r.extractSVGInnerContent(iconPath)
		if err != nil {
			// Skip icons that can't be read, but continue with others
			continue
		}

		// Use <symbol> instead of <g> for better viewBox handling
		// <symbol> is designed for reusable SVG content
		if viewBox != "" {
			fmt.Fprintf(sb, `    <symbol id="%s" viewBox="%s">%s</symbol>`, iconID, viewBox, innerContent)
		} else {
			fmt.Fprintf(sb, `    <symbol id="%s">%s</symbol>`, iconID, innerContent)
		}
		sb.WriteString("\n")
	}

	sb.WriteString("  </defs>\n")
}

// extractSVGInnerContent reads an SVG file and extracts its inner content
// (everything between the outer <svg> tags, excluding the <svg> tags themselves)
// Returns: innerContent, viewBox, error
func (r SVGRenderer) extractSVGInnerContent(svgPath string) (string, string, error) {
	content, err := os.ReadFile(svgPath)
	if err != nil {
		return "", "", fmt.Errorf("failed to read SVG file %s: %w", svgPath, err)
	}

	decoder := xml.NewDecoder(strings.NewReader(string(content)))
	var viewBox string
	var innerContent strings.Builder
	depth := 0
	inSVG := false

	for {
		token, err := decoder.Token()
		if err != nil {
			if err == io.EOF {
				break
			}
			return "", "", fmt.Errorf("failed to parse SVG file %s: %w", svgPath, err)
		}

		switch t := token.(type) {
		case xml.StartElement:
			// Check if this is the root <svg> element
			if t.Name.Local == "svg" && depth == 0 {
				inSVG = true
				// Extract viewBox attribute
				for _, attr := range t.Attr {
					if attr.Name.Local == "viewBox" {
						viewBox = attr.Value
						break
					}
				}
				depth++
				continue
			}

			if inSVG {
				// Skip Inkscape and Sodipodi namespace elements
				if t.Name.Space == "http://www.inkscape.org/namespaces/inkscape" ||
					t.Name.Space == "http://sodipodi.sourceforge.net/DTD/sodipodi-0.dtd" {
					depth++
					continue
				}

				// Skip metadata elements
				if t.Name.Local == "metadata" || t.Name.Local == "namedview" {
					depth++
					continue
				}

				// Write the element
				innerContent.WriteString("<")
				innerContent.WriteString(t.Name.Local)

				// Write attributes (filtering out editor-specific ones)
				for _, attr := range t.Attr {
					// Skip Inkscape and Sodipodi attributes
					if attr.Name.Space == "http://www.inkscape.org/namespaces/inkscape" ||
						attr.Name.Space == "http://sodipodi.sourceforge.net/DTD/sodipodi-0.dtd" {
						continue
					}

					innerContent.WriteString(" ")
					if attr.Name.Space != "" {
						innerContent.WriteString(attr.Name.Space)
						innerContent.WriteString(":")
					}
					innerContent.WriteString(attr.Name.Local)
					innerContent.WriteString(`="`)
					innerContent.WriteString(escapeXMLAttr(attr.Value))
					innerContent.WriteString(`"`)
				}

				innerContent.WriteString(">")
				depth++
			}

		case xml.EndElement:
			if inSVG {
				// Skip Inkscape and Sodipodi namespace elements
				if t.Name.Space == "http://www.inkscape.org/namespaces/inkscape" ||
					t.Name.Space == "http://sodipodi.sourceforge.net/DTD/sodipodi-0.dtd" {
					depth--
					if depth == 0 {
						inSVG = false
					}
					continue
				}

				// Skip metadata elements
				if t.Name.Local == "metadata" || t.Name.Local == "namedview" {
					depth--
					if depth == 0 {
						inSVG = false
					}
					continue
				}

				// Check if we're closing the root <svg> element
				if t.Name.Local == "svg" && depth == 1 {
					inSVG = false
					depth--
					continue
				}

				// Write closing tag
				innerContent.WriteString("</")
				innerContent.WriteString(t.Name.Local)
				innerContent.WriteString(">")
				depth--

				if depth == 0 {
					inSVG = false
				}
			}

		case xml.CharData:
			if inSVG && depth > 0 {
				// Only include character data if we're inside a graphic element
				// and it's not just whitespace
				text := strings.TrimSpace(string(t))
				if text != "" {
					innerContent.WriteString(escapeXML(string(t)))
				}
			}

		case xml.Comment:
			// Skip comments
			continue
		}
	}

	return strings.TrimSpace(innerContent.String()), viewBox, nil
}

// escapeXMLAttr escapes XML attribute values
func escapeXMLAttr(s string) string {
	s = strings.ReplaceAll(s, "&", "&amp;")
	s = strings.ReplaceAll(s, "<", "&lt;")
	s = strings.ReplaceAll(s, ">", "&gt;")
	s = strings.ReplaceAll(s, `"`, "&quot;")
	return s
}

// wrapText breaks text into lines that fit within the specified width
// Uses a simple approximation: average character width ≈ fontSize * 0.6
func (r SVGRenderer) wrapText(text string, fontSize, maxWidth float64) []string {
	if text == "" {
		return []string{text}
	}

	// Approximate average character width (most fonts are roughly 0.6x the font size)
	avgCharWidth := fontSize * 0.5
	maxCharsPerLine := int(maxWidth / avgCharWidth)

	// If text fits on one line, return it as-is
	if len(text) <= maxCharsPerLine {
		return []string{text}
	}

	var lines []string
	words := strings.Fields(text)
	currentLine := ""

	for _, word := range words {
		testLine := currentLine
		if testLine != "" {
			testLine += " " + word
		} else {
			testLine = word
		}

		// Check if adding this word would exceed the line width
		if len(testLine) <= maxCharsPerLine {
			currentLine = testLine
		} else {
			// If current line has content, save it and start a new line
			if currentLine != "" {
				lines = append(lines, currentLine)
				currentLine = word
			} else {
				// Word is too long, break it (shouldn't happen often, but handle it)
				// Break the word itself if it's longer than maxCharsPerLine
				if len(word) > maxCharsPerLine {
					// Add what we have so far
					if currentLine != "" {
						lines = append(lines, currentLine)
					}
					// Break the long word
					for len(word) > maxCharsPerLine {
						lines = append(lines, word[:maxCharsPerLine])
						word = word[maxCharsPerLine:]
					}
					currentLine = word
				} else {
					currentLine = word
				}
			}
		}
	}

	// Add the last line if there's any remaining text
	if currentLine != "" {
		lines = append(lines, currentLine)
	}

	return lines
}

// escapeXML escapes XML special characters in text
func escapeXML(text string) string {
	text = strings.ReplaceAll(text, "&", "&amp;")
	text = strings.ReplaceAll(text, "<", "&lt;")
	text = strings.ReplaceAll(text, ">", "&gt;")
	text = strings.ReplaceAll(text, `"`, "&quot;")
	text = strings.ReplaceAll(text, "'", "&apos;")
	return text
}
