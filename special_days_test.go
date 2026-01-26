package galendar_test

import (
	"os"
	"testing"
	"time"

	"github.com/unkiwii/galendar"
)

func TestLoadSpecialDaysFromFile_ExpressionEvaluation_Year(t *testing.T) {
	tmpFile := createTempSpecialDaysFile(t, `date_format = "2/1"

[[day]]
when = "18/3"
text = "((year - 2011))º Aniversario De Casados"
icon = "assets/anniversary.svg"
`)
	defer os.Remove(tmpFile)

	cfg := galendar.Config{
		Year:  2024,
		Month: 3,
	}

	specialDays, err := galendar.LoadSpecialDaysFromFile(tmpFile, cfg)
	if err != nil {
		t.Fatalf("LoadSpecialDaysFromFile failed: %v", err)
	}

	// Test that year - 2011 evaluates to 13 for year 2024
	date := time.Date(2024, time.March, 18, 0, 0, 0, 0, time.UTC)
	day := specialDays.At(date)
	if day == nil {
		t.Fatalf("Expected to find special day for March 18, 2024")
	}
	expectedText := "13º Aniversario De Casados"
	if day.Note.Text != expectedText {
		t.Errorf("Expected text %q, got %q", expectedText, day.Note.Text)
	}
}

func TestLoadSpecialDaysFromFile_ExpressionEvaluation_Month(t *testing.T) {
	tmpFile := createTempSpecialDaysFile(t, `date_format = "2/1"

[[day]]
when = "18/3"
text = "Month ((month))"
icon = "assets/test.svg"
`)
	defer os.Remove(tmpFile)

	cfg := galendar.Config{
		Year:  2024,
		Month: 3,
	}

	specialDays, err := galendar.LoadSpecialDaysFromFile(tmpFile, cfg)
	if err != nil {
		t.Fatalf("LoadSpecialDaysFromFile failed: %v", err)
	}

	date := time.Date(2024, time.March, 18, 0, 0, 0, 0, time.UTC)
	day := specialDays.At(date)
	if day == nil {
		t.Fatalf("Expected to find special day for March 18, 2024")
	}
	expectedText := "Month 3"
	if day.Note.Text != expectedText {
		t.Errorf("Expected text %q, got %q", expectedText, day.Note.Text)
	}
}

func TestLoadSpecialDaysFromFile_ExpressionEvaluation_Day(t *testing.T) {
	tmpFile := createTempSpecialDaysFile(t, `date_format = "2/1"

[[day]]
when = "18/3"
text = "Day ((day))"
icon = "assets/test.svg"
`)
	defer os.Remove(tmpFile)

	cfg := galendar.Config{
		Year:  2024,
		Month: 3,
	}

	specialDays, err := galendar.LoadSpecialDaysFromFile(tmpFile, cfg)
	if err != nil {
		t.Fatalf("LoadSpecialDaysFromFile failed: %v", err)
	}

	date := time.Date(2024, time.March, 18, 0, 0, 0, 0, time.UTC)
	day := specialDays.At(date)
	if day == nil {
		t.Fatalf("Expected to find special day for March 18, 2024")
	}
	expectedText := "Day 18"
	if day.Note.Text != expectedText {
		t.Errorf("Expected text %q, got %q", expectedText, day.Note.Text)
	}
}

func TestLoadSpecialDaysFromFile_ExpressionEvaluation_CfgYear(t *testing.T) {
	tmpFile := createTempSpecialDaysFile(t, `date_format = "2/1"

[[day]]
when = "18/3"
text = "Config year: ((cfg.year))"
icon = "assets/test.svg"
`)
	defer os.Remove(tmpFile)

	cfg := galendar.Config{
		Year:  2024,
		Month: 3,
	}

	specialDays, err := galendar.LoadSpecialDaysFromFile(tmpFile, cfg)
	if err != nil {
		t.Fatalf("LoadSpecialDaysFromFile failed: %v", err)
	}

	date := time.Date(2024, time.March, 18, 0, 0, 0, 0, time.UTC)
	day := specialDays.At(date)
	if day == nil {
		t.Fatalf("Expected to find special day for March 18, 2024")
	}
	expectedText := "Config year: 2024"
	if day.Note.Text != expectedText {
		t.Errorf("Expected text %q, got %q", expectedText, day.Note.Text)
	}
}

func TestLoadSpecialDaysFromFile_ExpressionEvaluation_CfgMonth(t *testing.T) {
	tmpFile := createTempSpecialDaysFile(t, `date_format = "2/1"

[[day]]
when = "18/3"
text = "Config month: ((cfg.month))"
icon = "assets/test.svg"
`)
	defer os.Remove(tmpFile)

	cfg := galendar.Config{
		Year:  2024,
		Month: 3,
	}

	specialDays, err := galendar.LoadSpecialDaysFromFile(tmpFile, cfg)
	if err != nil {
		t.Fatalf("LoadSpecialDaysFromFile failed: %v", err)
	}

	date := time.Date(2024, time.March, 18, 0, 0, 0, 0, time.UTC)
	day := specialDays.At(date)
	if day == nil {
		t.Fatalf("Expected to find special day for March 18, 2024")
	}
	expectedText := "Config month: 3"
	if day.Note.Text != expectedText {
		t.Errorf("Expected text %q, got %q", expectedText, day.Note.Text)
	}
}

func TestLoadSpecialDaysFromFile_ExpressionEvaluation_MultipleExpressions(t *testing.T) {
	tmpFile := createTempSpecialDaysFile(t, `date_format = "2/1"

[[day]]
when = "18/3"
text = "Year ((year)) has ((year - 2000)) years since 2000"
icon = "assets/test.svg"
`)
	defer os.Remove(tmpFile)

	cfg := galendar.Config{
		Year:  2024,
		Month: 3,
	}

	specialDays, err := galendar.LoadSpecialDaysFromFile(tmpFile, cfg)
	if err != nil {
		t.Fatalf("LoadSpecialDaysFromFile failed: %v", err)
	}

	date := time.Date(2024, time.March, 18, 0, 0, 0, 0, time.UTC)
	day := specialDays.At(date)
	if day == nil {
		t.Fatalf("Expected to find special day for March 18, 2024")
	}
	expectedText := "Year 2024 has 24 years since 2000"
	if day.Note.Text != expectedText {
		t.Errorf("Expected text %q, got %q", expectedText, day.Note.Text)
	}
}

func TestLoadSpecialDaysFromFile_ExpressionEvaluation_ArithmeticOperations(t *testing.T) {
	tests := []struct {
		name     string
		expr     string
		expected string
	}{
		{
			name:     "addition",
			expr:     "year + 1",
			expected: "2025",
		},
		{
			name:     "subtraction",
			expr:     "year - 2011",
			expected: "13",
		},
		{
			name:     "multiple operations",
			expr:     "year - 2000 + 10",
			expected: "34",
		},
		{
			name:     "month addition",
			expr:     "month + 1",
			expected: "4",
		},
		{
			name:     "day subtraction",
			expr:     "day - 1",
			expected: "17",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpFile := createTempSpecialDaysFile(t, `date_format = "2/1"

[[day]]
when = "18/3"
text = "((`+tt.expr+`))"
icon = "assets/test.svg"
`)
			defer os.Remove(tmpFile)

			cfg := galendar.Config{
				Year:  2024,
				Month: 3,
			}

			specialDays, err := galendar.LoadSpecialDaysFromFile(tmpFile, cfg)
			if err != nil {
				t.Fatalf("LoadSpecialDaysFromFile failed: %v", err)
			}

			date := time.Date(2024, time.March, 18, 0, 0, 0, 0, time.UTC)
			day := specialDays.At(date)
			if day == nil {
				t.Fatalf("Expected to find special day for March 18, 2024")
			}
			if day.Note.Text != tt.expected {
				t.Errorf("Expected text %q, got %q", tt.expected, day.Note.Text)
			}
		})
	}
}

func TestLoadSpecialDaysFromFile_SkipInvalidExpressions(t *testing.T) {
	tmpFile := createTempSpecialDaysFile(t, `date_format = "2/1"

[[day]]
when = "18/3"
text = "((year - 2011))º Aniversario De Casados"
icon = "assets/anniversary.svg"
`)
	defer os.Remove(tmpFile)

	// Use year 2010, which makes year - 2011 = -1 (should be skipped)
	cfg := galendar.Config{
		Year:  2010,
		Month: 3,
	}

	specialDays, err := galendar.LoadSpecialDaysFromFile(tmpFile, cfg)
	if err != nil {
		t.Fatalf("LoadSpecialDaysFromFile failed: %v", err)
	}

	// The day with "((year - 2011))" should be skipped when year is 2010
	date := time.Date(2010, time.March, 18, 0, 0, 0, 0, time.UTC)
	day := specialDays.At(date)
	if day != nil {
		t.Errorf("Expected special day to be skipped when expression evaluates to ≤ 0, but found: %+v", day)
	}
}

func TestLoadSpecialDaysFromFile_ExpressionEvaluation_IconProperty(t *testing.T) {
	tmpFile := createTempSpecialDaysFile(t, `date_format = "2/1"

[[day]]
when = "18/3"
text = "Test"
icon = "assets/icon-((year)).svg"
`)
	defer os.Remove(tmpFile)

	cfg := galendar.Config{
		Year:  2024,
		Month: 3,
	}

	specialDays, err := galendar.LoadSpecialDaysFromFile(tmpFile, cfg)
	if err != nil {
		t.Fatalf("LoadSpecialDaysFromFile failed: %v", err)
	}

	date := time.Date(2024, time.March, 18, 0, 0, 0, 0, time.UTC)
	day := specialDays.At(date)
	if day == nil {
		t.Fatalf("Expected to find special day for March 18, 2024")
	}
	expectedIcon := "assets/icon-2024.svg"
	if day.Icon != expectedIcon {
		t.Errorf("Expected icon %q, got %q", expectedIcon, day.Icon)
	}
}

func TestLoadSpecialDaysFromFile_ExpressionEvaluation_FontProperty(t *testing.T) {
	tmpFile := createTempSpecialDaysFile(t, `date_format = "2/1"

[[day]]
when = "18/3"
text = "Test"
font = "font-((month))"
`)
	defer os.Remove(tmpFile)

	cfg := galendar.Config{
		Year:  2024,
		Month: 3,
	}

	specialDays, err := galendar.LoadSpecialDaysFromFile(tmpFile, cfg)
	if err != nil {
		t.Fatalf("LoadSpecialDaysFromFile failed: %v", err)
	}

	date := time.Date(2024, time.March, 18, 0, 0, 0, 0, time.UTC)
	day := specialDays.At(date)
	if day == nil {
		t.Fatalf("Expected to find special day for March 18, 2024")
	}
	expectedFont := "font-3"
	if day.Note.Font != expectedFont {
		t.Errorf("Expected font %q, got %q", expectedFont, day.Note.Font)
	}
}

func TestLoadSpecialDaysFromFile_EmptyFile(t *testing.T) {
	cfg := galendar.Config{
		Year:  2024,
		Month: 3,
	}

	specialDays, err := galendar.LoadSpecialDaysFromFile("", cfg)
	if err != nil {
		t.Fatalf("LoadSpecialDaysFromFile should not error on empty filename: %v", err)
	}
	if specialDays != nil {
		t.Errorf("Expected nil for empty filename, got %v", specialDays)
	}
}

func TestLoadSpecialDaysFromFile_RelativeDate_ThirdSunday(t *testing.T) {
	tmpFile := createTempSpecialDaysFile(t, `date_format = "2/1"

[[day]]
when = "((3rd sunday))/10"
text = "Mother's day"
icon = "assets/mothersday.svg"
`)
	defer os.Remove(tmpFile)

	cfg := galendar.Config{
		Year:  2024,
		Month: 10,
	}

	specialDays, err := galendar.LoadSpecialDaysFromFile(tmpFile, cfg)
	if err != nil {
		t.Fatalf("LoadSpecialDaysFromFile failed: %v", err)
	}

	// Third Sunday of October 2024 is October 20
	date := time.Date(2024, time.October, 20, 0, 0, 0, 0, time.UTC)
	day := specialDays.At(date)
	if day == nil {
		t.Fatalf("Expected to find special day for October 20, 2024 (3rd Sunday)")
	}
	if day.Note.Text != "Mother's day" {
		t.Errorf("Expected text %q, got %q", "Mother's day", day.Note.Text)
	}
}

func TestLoadSpecialDaysFromFile_RelativeDate_FirstFriday(t *testing.T) {
	tmpFile := createTempSpecialDaysFile(t, `date_format = "2/1"

[[day]]
when = "((1st friday))/3"
text = "First Friday"
icon = "assets/test.svg"
`)
	defer os.Remove(tmpFile)

	cfg := galendar.Config{
		Year:  2024,
		Month: 3,
	}

	specialDays, err := galendar.LoadSpecialDaysFromFile(tmpFile, cfg)
	if err != nil {
		t.Fatalf("LoadSpecialDaysFromFile failed: %v", err)
	}

	// First Friday of March 2024 is March 1
	date := time.Date(2024, time.March, 1, 0, 0, 0, 0, time.UTC)
	day := specialDays.At(date)
	if day == nil {
		t.Fatalf("Expected to find special day for March 1, 2024 (1st Friday)")
	}
	if day.Note.Text != "First Friday" {
		t.Errorf("Expected text %q, got %q", "First Friday", day.Note.Text)
	}
}

func TestLoadSpecialDaysFromFile_RelativeDate_LastMonday(t *testing.T) {
	tmpFile := createTempSpecialDaysFile(t, `date_format = "2/1"

[[day]]
when = "((last monday))/3"
text = "Last Monday"
icon = "assets/test.svg"
`)
	defer os.Remove(tmpFile)

	cfg := galendar.Config{
		Year:  2024,
		Month: 3,
	}

	specialDays, err := galendar.LoadSpecialDaysFromFile(tmpFile, cfg)
	if err != nil {
		t.Fatalf("LoadSpecialDaysFromFile failed: %v", err)
	}

	// Last Monday of March 2024 is March 25
	date := time.Date(2024, time.March, 25, 0, 0, 0, 0, time.UTC)
	day := specialDays.At(date)
	if day == nil {
		t.Fatalf("Expected to find special day for March 25, 2024 (last Monday)")
	}
	if day.Note.Text != "Last Monday" {
		t.Errorf("Expected text %q, got %q", "Last Monday", day.Note.Text)
	}
}

func TestLoadSpecialDaysFromFile_RelativeDate_WithExpressions(t *testing.T) {
	tmpFile := createTempSpecialDaysFile(t, `date_format = "2/1"

[[day]]
when = "((2nd sunday))/5"
text = "Year ((year)) - Week ((year - 2000))"
icon = "assets/test.svg"
`)
	defer os.Remove(tmpFile)

	cfg := galendar.Config{
		Year:  2024,
		Month: 5,
	}

	specialDays, err := galendar.LoadSpecialDaysFromFile(tmpFile, cfg)
	if err != nil {
		t.Fatalf("LoadSpecialDaysFromFile failed: %v", err)
	}

	// Second Sunday of May 2024 is May 12
	date := time.Date(2024, time.May, 12, 0, 0, 0, 0, time.UTC)
	day := specialDays.At(date)
	if day == nil {
		t.Fatalf("Expected to find special day for May 12, 2024 (2nd Sunday)")
	}
	expectedText := "Year 2024 - Week 24"
	if day.Note.Text != expectedText {
		t.Errorf("Expected text %q, got %q", expectedText, day.Note.Text)
	}
}

func createTempSpecialDaysFile(t *testing.T, content string) string {
	tmpFile, err := os.CreateTemp("", "special_days_*.toml")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer tmpFile.Close()

	if _, err := tmpFile.WriteString(content); err != nil {
		t.Fatalf("Failed to write temp file: %v", err)
	}

	return tmpFile.Name()
}
