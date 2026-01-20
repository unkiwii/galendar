package galendar

import (
	"fmt"
	"time"
)

// TODO: change Calendar name to Month
// Calendar represents a calendar for a given month
type Calendar struct {
	Year      int
	Month     int
	Weeks     [][]Day
	WeekStart time.Weekday
}

// Day represents a single day in the calendar
type Day struct {
	Date           time.Time
	DayNumber      int
	IsCurrentMonth bool
}

// NewCalendar creates a new calendar for the given month and year
func NewCalendar(year, month int, weekStart time.Weekday) (*Calendar, error) {
	if month < 1 || month > 12 {
		return nil, fmt.Errorf("invalid month: %d (must be 1-12)", month)
	}

	cal := &Calendar{
		Year:      year,
		Month:     month,
		WeekStart: weekStart,
	}

	// Get the first day of the month
	firstDay := time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.UTC)

	// Get the last day of the month
	lastDay := firstDay.AddDate(0, 1, -1)

	// Determine the starting day of the week for the calendar
	// Go's time.Weekday: Sunday=0, Monday=1, ..., Saturday=6
	firstWeekday := int(firstDay.Weekday())

	// Convert to our week start system (0=Sunday, 1=Monday, ..., 6=Saturday)
	// If weekStart is 0 (Sunday), firstWeekday (0) maps to 0
	// If weekStart is 1 (Monday), firstWeekday (1) maps to 0, etc.
	startOffset := (firstWeekday - int(weekStart) + 7) % 7

	// Calculate how many days we need to show before the first day
	daysBefore := startOffset

	// Start from the first day we need to show
	startDate := firstDay.AddDate(0, 0, -daysBefore)

	// Build the calendar grid (6 weeks × 7 days = 42 days max)
	var weeks [][]Day
	currentDate := startDate

	for range 6 {
		var weekDays []Day
		for day := range 7 {
			isCurrentMonth := currentDate.Month() == time.Month(month) && currentDate.Year() == year

			weekDays = append(weekDays, Day{
				Date:           currentDate,
				DayNumber:      currentDate.Day(),
				IsCurrentMonth: isCurrentMonth,
			})

			currentDate = currentDate.AddDate(0, 0, 1)

			// Stop if we've passed the last day and we're starting a new week
			if currentDate.After(lastDay) && day == 6 {
				break
			}
		}

		weeks = append(weeks, weekDays)

		// Stop if we've passed the last day
		if currentDate.After(lastDay) {
			break
		}
	}

	cal.Weeks = weeks
	return cal, nil
}
