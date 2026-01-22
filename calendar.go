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

type Note struct {
	Text string
	Font string
	Size float64
}

// Day represents a single day in the calendar
type Day struct {
	Date           time.Time
	DayNumber      int
	IsCurrentMonth bool
	HolidayMark    bool
	Icon           string // TODO: maybe svg image?
	Note           *Note
}

func (day Day) TextColor() (r, g, b, a int) {
	if !day.IsCurrentMonth {
		return 200, 200, 200, 0
	}

	return 0, 0, 0, 1
}

func (day Day) FillColor() (r, g, b, a int) {
	if !day.IsCurrentMonth {
		return 0, 0, 0, 0
	}

	if day.IsHoliday() {
		return 240, 240, 240, 1
	}

	return 0, 0, 0, 0
}

func (day Day) IsHoliday() bool {
	weekday := day.Date.Weekday()
	return day.HolidayMark || weekday == time.Saturday || weekday == time.Sunday
}

func (day Day) Name() string {
	return day.Date.Format(time.DateOnly)
}

type md struct {
	m int
	d int
}

type specialDay struct {
	holiday bool
	icon    string
	note    Note
}

// TODO: move this to a file
var specialDays = map[md]specialDay{
	// TODO: pdf doesn't support ñ apparently
	{m: 1, d: 1}: {
		holiday: true,
		icon:    "",
		note:    Note{Text: "Año Nuevo", Font: "/usr/share/fonts/truetype/AcPlus_IBM_VGA_9x16.ttf", Size: 22},
	},
	{m: 1, d: 23}: {
		holiday: false,
		icon:    "Birthday",
		note:    Note{Text: "Mora"},
	},
	{m: 1, d: 25}: {
		holiday: false,
		icon:    "Birthday",
		note:    Note{Text: "Lucas"},
	},
	{m: 2, d: 23}: {
		holiday: false,
		icon:    "Birthday",
		note:    Note{Text: "Malena"},
	},
	{m: 5, d: 1}: {
		holiday: true,
		icon:    "",
		note:    Note{Text: "Dia del Trabajador", Size: 14},
	},
	{m: 7, d: 9}: {
		holiday: true,
		icon:    "",
		note:    Note{Text: "Dia de la Independencia", Size: 14},
	},
	{m: 12, d: 25}: {
		holiday: true,
		icon:    "",
		note:    Note{Text: "Navidad"},
	},
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
			}

			if special, ok := specialDays[md{m: int(currentDate.Month()), d: currentDate.Day()}]; ok {
				weekDay.HolidayMark = special.holiday
				weekDay.Icon = special.icon
				weekDay.Note = &special.note
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
