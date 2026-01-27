package galendar

import (
	"fmt"
	"time"
)

// TODO: change Calendar name to Month
// Calendar represents a calendar for a given month
type Calendar struct {
	Year        int
	Month       int
	Weeks       [][]Day
	WeekStart   time.Weekday
	SpecialDays SpecialDays
}

// NewCalendar creates a new calendar for the given month and year
func NewCalendar(year, month int, weekStart time.Weekday, specialDays SpecialDays) (Calendar, error) {
	var cal Calendar

	if month < 1 || month > 12 {
		return cal, fmt.Errorf("invalid month: %d (must be 1-12)", month)
	}

	cal.Year = year
	cal.Month = month
	cal.WeekStart = weekStart
	cal.SpecialDays = specialDays

	firstDayOfMonth := time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.UTC)
	lastDayOfMonth := firstDayOfMonth.AddDate(0, 1, -1)
	firstWeekday := int(firstDayOfMonth.Weekday())

	// Convert to our week start system (0=Sunday, 1=Monday, ..., 6=Saturday)
	// If weekStart is 0 (Sunday), firstWeekday (0) maps to 0
	// If weekStart is 1 (Monday), firstWeekday (1) maps to 0, etc.
	startOffset := (firstWeekday - int(weekStart) + 7) % 7

	// Start from the first day we need to show
	startDate := firstDayOfMonth.AddDate(0, 0, -startOffset)

	// Build the calendar grid (6 weeks × 7 days = 42 days max)
	var weeks [][]Day
	currentDate := startDate

	for range 6 {
		var weekDays []Day
		for day := range 7 {
			isCurrentMonth := currentDate.Month() == time.Month(month) && currentDate.Year() == year

			weekDay := Day{
				Date:           currentDate,
				DayNumber:      currentDate.Day(),
				IsCurrentMonth: isCurrentMonth,
				special:        specialDays.At(currentDate),
			}

			weekDays = append(weekDays, weekDay)

			currentDate = currentDate.AddDate(0, 0, 1)

			// Stop if we've passed the last day and we're starting a new week
			if currentDate.After(lastDayOfMonth) && day == 6 {
				break
			}
		}

		weeks = append(weeks, weekDays)

		// Stop if we've passed the last day
		if currentDate.After(lastDayOfMonth) {
			break
		}
	}

	cal.Weeks = weeks
	return cal, nil
}

func (cal Calendar) CloneAt(month int) (Calendar, error) {
	return NewCalendar(cal.Year, month, cal.WeekStart, cal.SpecialDays)
}

type Day struct {
	Date           time.Time
	DayNumber      int
	IsCurrentMonth bool
	special        *SpecialDay
}

func (day Day) TextColor() (r, g, b, a int) {
	if !day.IsCurrentMonth {
		return 128, 128, 128, 0
	}

	return 0, 0, 0, 1
}

func (day Day) FillColor() (r, g, b, a int) {
	if !day.IsCurrentMonth {
		return 0, 0, 0, 0
	}

	if day.IsHoliday() {
		return 200, 200, 200, 1
	}

	return 0, 0, 0, 0
}

func (day Day) IsHoliday() bool {
	weekday := day.Date.Weekday()
	if weekday == time.Saturday || weekday == time.Sunday {
		return true
	}

	if day.special != nil {
		return day.special.Holiday
	}

	return false
}

func (day Day) Name() string {
	return day.Date.Format(time.DateOnly)
}

func (day Day) Note() *SpecialDayNote {
	if day.special == nil {
		return nil
	}

	return &day.special.Note
}
