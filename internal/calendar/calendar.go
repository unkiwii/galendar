package calendar

import (
	"fmt"
	"time"

	"github.com/unkiwii/galendar/internal/config"
)

// Calendar represents a calendar for a given month
type Calendar struct {
	Year      int
	Month     int
	MonthName string
	Weeks     [][]Day
	WeekStart config.WeekStart
}

// Day represents a single day in the calendar
type Day struct {
	Date           time.Time
	DayNumber      int
	IsCurrentMonth bool
}

// NewCalendar creates a new calendar for the given month and year
func NewCalendar(year, month int, weekStart config.WeekStart) (*Calendar, error) {
	if month < 1 || month > 12 {
		return nil, fmt.Errorf("invalid month: %d (must be 1-12)", month)
	}

	cal := &Calendar{
		Year:      year,
		Month:     month,
		MonthName: time.Month(month).String(),
		WeekStart: weekStart,
	}

	// Get the first day of the month
	firstDay := time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.UTC)

	// Get the last day of the month
	lastDay := firstDay.AddDate(0, 1, -1)

	// Determine the starting day of the week for the calendar
	var startOffset int
	firstWeekday := int(firstDay.Weekday())

	if weekStart == config.Sunday {
		// Sunday = 0, so firstWeekday is already correct
		startOffset = firstWeekday
	} else {
		// Monday = 0, so we need to adjust
		startOffset = (firstWeekday + 6) % 7 // Convert Sunday=0 to Monday=0
	}

	// Calculate how many days we need to show before the first day
	daysBefore := startOffset

	// Start from the first day we need to show
	startDate := firstDay.AddDate(0, 0, -daysBefore)

	// Build the calendar grid (6 weeks × 7 days = 42 days max)
	var weeks [][]Day
	currentDate := startDate

	for week := 0; week < 6; week++ {
		var weekDays []Day
		for day := 0; day < 7; day++ {
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

// GetWeekdayNames returns the names of weekdays based on week start
func GetWeekdayNames(weekStart config.WeekStart) []string {
	names := []string{"Sunday", "Monday", "Tuesday", "Wednesday", "Thursday", "Friday", "Saturday"}

	if weekStart == config.Monday {
		// Rotate to start with Monday
		return append(names[1:], names[0])
	}

	return names
}

// GetWeekdayAbbreviations returns abbreviated weekday names
func GetWeekdayAbbreviations(weekStart config.WeekStart) []string {
	names := []string{"Sun", "Mon", "Tue", "Wed", "Thu", "Fri", "Sat"}

	if weekStart == config.Monday {
		return append(names[1:], names[0])
	}

	return names
}
