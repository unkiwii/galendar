package galendar

import (
	"fmt"
	"time"

	"github.com/BurntSushi/toml"
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

func LoadSpecialDaysFromFile(filename string) (SpecialDays, error) {
	if filename == "" {
		return nil, nil
	}

	var specialDaysFile specialDaysFile

	_, err := toml.DecodeFile(filename, &specialDaysFile)
	if err != nil {
		return nil, fmt.Errorf("can't decode toml file %q: %w", filename, err)
	}

	days := SpecialDays{}
	for _, day := range specialDaysFile.Day {
		key, err := specialDaysKeyFromString(specialDaysFile.DateFormat, day.When)
		if err != nil {
			return nil, fmt.Errorf("invalid 'when' value %q: %w", day.When, err)
		}
		days[key] = SpecialDay{
			Holiday: day.Holiday,
			Icon:    day.Icon,
			Note: SpecialDayNote{
				Text: day.Text,
				Font: day.Font,
				Size: day.Size,
			},
		}
	}

	return days, nil
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

type specialDaysFile struct {
	DateFormat string `toml:"date_format"`
	Day        []struct {
		When    string
		Holiday bool
		Icon    string
		Text    string
		Font    string
		Size    float64
	}
}

type specialDaysKey struct {
	month int
	day   int
}

func (key specialDaysKey) String() string {
	return fmt.Sprintf("%d/%d", key.month, key.day)
}

func specialDaysKeyFromString(layout, s string) (specialDaysKey, error) {
	t, err := time.Parse(layout, s)
	if err != nil {
		return specialDaysKey{}, fmt.Errorf("can't parse %q as %q: %w", s, layout, err)
	}

	return specialDaysKeyFromTime(t), nil
}

func specialDaysKeyFromTime(t time.Time) specialDaysKey {
	return specialDaysKey{
		month: int(t.Month()),
		day:   t.Day(),
	}
}
