package mutingschedule

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// weekdays are the only weekday values the API accepts. It takes
// neither abbreviations nor ranges, unlike Alertmanager's own time
// intervals, so the set is checked here rather than left to a 400.
var weekdays = map[string]struct{}{
	"monday":    {},
	"tuesday":   {},
	"wednesday": {},
	"thursday":  {},
	"friday":    {},
	"saturday":  {},
	"sunday":    {},
}

// months maps the month names the API accepts to their number.
var months = map[string]int{
	"january":   1,
	"february":  2,
	"march":     3,
	"april":     4,
	"may":       5,
	"june":      6,
	"july":      7,
	"august":    8,
	"september": 9,
	"october":   10,
	"november":  11,
	"december":  12,
}

var timeOfDayPattern = regexp.MustCompile(`^([01][0-9]|2[0-3]):[0-5][0-9]$`)

// validateWeekday checks one entry of an interval's weekdays.
func validateWeekday(value string) error {
	if _, ok := weekdays[value]; !ok {
		return fmt.Errorf(
			"%q is not a weekday: use a lowercase day name such as "+
				"\"monday\". Ranges and abbreviations are not accepted",
			value,
		)
	}

	return nil
}

// validateDayOfMonth checks one entry of an interval's days_of_month.
// An entry is a day or an inclusive "begin:end" range. Days count
// from the start of the month, or from its end when negative, so -1
// is the last day of whatever month the schedule lands in.
func validateDayOfMonth(value string) error {
	begin, end, err := parseRange(value, func(part string) (int, error) {
		day, err := strconv.Atoi(part)
		if err != nil {
			return 0, fmt.Errorf("%q is not a number", part)
		}

		if day == 0 || day < -31 || day > 31 {
			return 0, fmt.Errorf(
				"%d is not a day of the month: days run 1 to 31, or -31 "+
					"to -1 counting back from the end of the month",
				day,
			)
		}

		return day, nil
	})
	if err != nil {
		return err
	}

	// A range that mixes the two directions has no fixed length, so
	// the API cannot resolve it.
	if begin < 0 && end > 0 {
		return fmt.Errorf(
			"%q counts from both ends of the month: if the first day is "+
				"negative the last one must be too",
			value,
		)
	}

	// Negative days are clamped against the shortest month, so this
	// only catches ranges that are empty in every month.
	if clampDay(begin) > clampDay(end) {
		return fmt.Errorf(
			"%q ends before it starts", value,
		)
	}

	return nil
}

// clampDay resolves a negative day against a 28-day month, the
// shortest one a range has to hold for.
func clampDay(day int) int {
	if day < 0 {
		return 28 + day
	}

	return day
}

// validateMonth checks one entry of an interval's months. An entry is
// a month name or number, or an inclusive range of either.
func validateMonth(value string) error {
	begin, end, err := parseRange(value, func(part string) (int, error) {
		if month, ok := months[part]; ok {
			return month, nil
		}

		month, err := strconv.Atoi(part)
		if err != nil || month < 1 || month > 12 {
			return 0, fmt.Errorf(
				"%q is not a month: use a lowercase month name such as "+
					"\"january\", or a number from 1 to 12",
				part,
			)
		}

		return month, nil
	})
	if err != nil {
		return err
	}

	if begin > end {
		return fmt.Errorf("%q ends before it starts", value)
	}

	return nil
}

// validateYear checks one entry of an interval's years. An entry is a
// year or an inclusive range of years.
func validateYear(value string) error {
	begin, end, err := parseRange(value, func(part string) (int, error) {
		year, err := strconv.Atoi(part)
		if err != nil {
			return 0, fmt.Errorf("%q is not a year", part)
		}

		return year, nil
	})
	if err != nil {
		return err
	}

	if begin > end {
		return fmt.Errorf("%q ends before it starts", value)
	}

	return nil
}

// validateLocation checks an interval's timezone. The API answers an
// unknown one with a bare 500, so it is worth resolving up front.
func validateLocation(value string) error {
	if _, err := time.LoadLocation(value); err != nil {
		return fmt.Errorf(
			"%q is not an IANA timezone, such as \"America/New_York\" "+
				"or \"UTC\"",
			value,
		)
	}

	return nil
}

// validateTimeOfDay checks one end of a time range.
func validateTimeOfDay(value string) error {
	if !timeOfDayPattern.MatchString(value) {
		return fmt.Errorf(
			"%q is not a time of day: use 24-hour HH:MM, from \"00:00\" "+
				"to \"23:59\"",
			value,
		)
	}

	return nil
}

// parseRange splits an entry that is either a single value or an
// inclusive "begin:end" range, and parses both ends with parse.
func parseRange(
	value string,
	parse func(string) (int, error),
) (int, int, error) {
	begin, end, isRange := strings.Cut(value, ":")
	if !isRange {
		parsed, err := parse(value)

		return parsed, parsed, err
	}

	parsedBegin, err := parse(begin)
	if err != nil {
		return 0, 0, err
	}

	parsedEnd, err := parse(end)
	if err != nil {
		return 0, 0, err
	}

	return parsedBegin, parsedEnd, nil
}
