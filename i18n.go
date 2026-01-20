package galendar

import (
	"time"
)

type Language string

const (
	Spanish = Language("es")
	English = Language("en")
)

var i18nStrings map[Language]map[string]string

func init() {
	// TODO: move this to a file and load it here
	i18nStrings = map[Language]map[string]string{}

	i18nStrings[English] = map[string]string{
		"Sunday":    "Sunday",
		"Sun":       "Sun",
		"Monday":    "Monday",
		"Mon":       "Mon",
		"Tuesday":   "Tuesday",
		"Tue":       "Tue",
		"Wednesday": "Wednesday",
		"Wed":       "Wed",
		"Thursday":  "Thursday",
		"Thu":       "Thu",
		"Friday":    "Friday",
		"Fri":       "Fri",
		"Saturday":  "Saturday",
		"Sat":       "Sat",
		"January":   "January",
		"February":  "February",
		"March":     "March",
		"April":     "April",
		"May":       "May",
		"June":      "June",
		"July":      "July",
		"August":    "August",
		"September": "September",
		"October":   "October",
		"November":  "November",
		"December":  "December",
		"calendar":  "calendar",
	}

	i18nStrings[Spanish] = map[string]string{
		"Sunday":    "Domingo",
		"Sun":       "D",
		"Monday":    "Lunes",
		"Mon":       "L",
		"Tuesday":   "Martes",
		"Tue":       "M",
		"Wednesday": "Miércoles",
		"Wed":       "M",
		"Thursday":  "Jueves",
		"Thu":       "J",
		"Friday":    "Viernes",
		"Fri":       "V",
		"Saturday":  "Sábado",
		"Sat":       "S",
		"January":   "Enero",
		"February":  "Febrero",
		"March":     "Marzo",
		"April":     "Abril",
		"May":       "Mayo",
		"June":      "Junio",
		"July":      "Julio",
		"August":    "Agosto",
		"September": "Septiembre",
		"October":   "Octubre",
		"November":  "Noviembre",
		"December":  "Diciembre",
		"calendar":  "calendar",
	}
}

func (lang Language) MonthName(month int) string {
	return lang.Read(time.Month(month).String())
}

func (lang Language) WeekdayAbbreviations(weekStart time.Weekday) []string {
	names := []string{"Sun", "Mon", "Tue", "Wed", "Thu", "Fri", "Sat"}
	translated := make([]string, len(names))
	for i, name := range names {
		translated[i] = lang.Read(name)
	}
	return rotateWeekdays(translated, int(weekStart))
}

func (lang Language) Read(key string) (ret string) {
	strings, ok := i18nStrings[lang]
	if !ok {
		return key
	}
	val, ok := strings[key]
	if !ok {
		return key
	}
	return val
}

func IsValidLanguage(lang Language) bool {
	_, ok := i18nStrings[lang]
	return ok
}

func rotateWeekdays(names []string, startDay int) []string {
	if startDay < 0 || startDay > 6 {
		startDay = 0
	}
	if startDay == 0 {
		return names
	}
	return append(names[startDay:], names[:startDay]...)
}
