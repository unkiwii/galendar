package galendar

import (
	"time"
)

type SpecialDay struct {
	Date    time.Time
	Holiday bool
	Icon    string
	Note    SpecialDayNote
}

type SpecialDayNote struct {
	Text string
	Font string
	Size float64
}

type SpecialDays map[specialDaysKey]SpecialDay

func (days SpecialDays) At(date time.Time) *SpecialDay {
	if len(days) == 0 {
		return nil
	}
	key := specialDaysKeyFromTime(date)
	if day, ok := days[key]; ok {
		return &day
	}
	return nil
}
