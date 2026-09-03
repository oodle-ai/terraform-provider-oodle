package mutingschedule

import (
	"testing"

	"github.com/rubrikinc/testwell/assert"
)

func TestValidateWeekday(t *testing.T) {
	for _, value := range []string{"monday", "sunday", "wednesday"} {
		assert.Nil(t, validateWeekday(value), "value: %v", value)
	}

	// The API takes neither the abbreviations nor the ranges
	// Alertmanager's own time intervals accept, so neither may pass.
	for _, value := range []string{
		"mon", "Monday", "MONDAY", "1", "monday:friday", "funday", "",
	} {
		assert.NotNil(t, validateWeekday(value), "value: %v", value)
	}
}

func TestValidateDayOfMonth(t *testing.T) {
	for _, value := range []string{
		"1", "31", "-1", "-31", "1:5", "-7:-1", "15:15",
	} {
		assert.Nil(t, validateDayOfMonth(value), "value: %v", value)
	}

	for _, value := range []string{
		"0",     // no zeroth day
		"32",    // past the longest month
		"-32",   // past the longest month, counting back
		"5:1",   // ends before it starts
		"-1:-7", // ends before it starts, counting back
		"-7:5",  // counts from both ends
		"first",
		"",
	} {
		assert.NotNil(t, validateDayOfMonth(value), "value: %v", value)
	}
}

// TestValidateDayOfMonthAllowsRangesShortMonthsClamp pins that a
// range only a long month can hold in full is still accepted: the
// API clamps it rather than rejecting it.
func TestValidateDayOfMonthAllowsRangesShortMonthsClamp(t *testing.T) {
	assert.Nil(t, validateDayOfMonth("28:31"))
	assert.Nil(t, validateDayOfMonth("-3:-1"))
}

func TestValidateMonth(t *testing.T) {
	for _, value := range []string{
		"january", "december", "1", "12", "1:3", "january:march", "3:3",
	} {
		assert.Nil(t, validateMonth(value), "value: %v", value)
	}

	for _, value := range []string{
		"jan",     // the API rejects abbreviations
		"January", // and capitals
		"0", "13", // out of range
		"3:1",           // ends before it starts
		"march:january", // ends before it starts, by name
		"smarch",
		"",
	} {
		assert.NotNil(t, validateMonth(value), "value: %v", value)
	}
}

func TestValidateYear(t *testing.T) {
	for _, value := range []string{"2026", "2026:2030", "2026:2026"} {
		assert.Nil(t, validateYear(value), "value: %v", value)
	}

	for _, value := range []string{"2030:2026", "abcd", "20x6", ""} {
		assert.NotNil(t, validateYear(value), "value: %v", value)
	}
}

// TestValidateLocation covers the check that earns its keep: the API
// answers an unknown timezone with a bare 500.
func TestValidateLocation(t *testing.T) {
	for _, value := range []string{
		"UTC", "America/New_York", "Asia/Kolkata", "Europe/London",
	} {
		assert.Nil(t, validateLocation(value), "value: %v", value)
	}

	for _, value := range []string{"Mars/Olympus", "EST5EDT/nope", "PST"} {
		assert.NotNil(t, validateLocation(value), "value: %v", value)
	}
}

func TestValidateTimeOfDay(t *testing.T) {
	for _, value := range []string{"00:00", "09:30", "23:59"} {
		assert.Nil(t, validateTimeOfDay(value), "value: %v", value)
	}

	// 24:00 reads like the end of the day but the API rejects it.
	for _, value := range []string{
		"24:00", "9:30", "09:60", "0900", "09:30:00", "",
	} {
		assert.NotNil(t, validateTimeOfDay(value), "value: %v", value)
	}
}
