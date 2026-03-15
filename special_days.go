package galendar

import (
	"fmt"
	"strings"
	"time"
)

type SpecialDay struct {
	Date    time.Time
	Holiday bool
	Icon    string
	Note    SpecialDayNote
}

func (day SpecialDay) Merge(other SpecialDay) SpecialDay {
	if day.Date.Compare(other.Date) != 0 {
		return day
	}

	icon := day.Icon
	if icon == "" {
		icon = other.Icon
	}

	var noteText strings.Builder
	if day.Note.Text != "" {
		noteText.WriteString(day.Note.Text)
	}
	if other.Note.Text != "" {
		if noteText.Len() != 0 {
			noteText.WriteString(", ")
		}
		noteText.WriteString(other.Note.Text)
	}

	return SpecialDay{
		Date:    day.Date,
		Holiday: day.Holiday || other.Holiday,
		Icon:    icon,
		Note: SpecialDayNote{
			Text: noteText.String(),
			Font: day.Note.Font,
			Size: day.Note.Size,
		},
	}
}

type SpecialDayNote struct {
	Text string
	Font string
	Size float64
}

type specialDaysKey struct {
	month int
	day   int
}

func (key specialDaysKey) String() string {
	return fmt.Sprintf("%d/%d", key.month, key.day)
}

type SpecialDays map[specialDaysKey]SpecialDay

func (days SpecialDays) String() string {
	var sb strings.Builder

	for k, v := range days {
		fmt.Fprintf(&sb, "  [%s] = %#v\n", k, v)
	}

	return sb.String()
}

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
