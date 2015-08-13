/*
Package arrow provides C-style date formating and parsing, along with other date goodies.

See the github project page at http://github.com/bmuller/arrow for more info.
*/
package arrow

import (
	"strconv"
	"strings"
	"time"
)

type Arrow struct {
	time.Time
}

// Like time's constants, but with Day and Week
const (
	Nanosecond  time.Duration = 1
	Microsecond               = 1000 * Nanosecond
	Millisecond               = 1000 * Microsecond
	Second                    = 1000 * Millisecond
	Minute                    = 60 * Second
	Hour                      = 60 * Minute
	Day                       = 24 * Hour
	Week                      = 7 * Day
)

func New(t time.Time) Arrow {
	return Arrow{t}
}

func UTC() Arrow {
	return New(time.Now().UTC())
}

func Unix(sec int64, nsec int64) Arrow {
	return New(time.Unix(sec, nsec))
}

func Epoch() Arrow {
	return Unix(0, 0)
}

func Now() Arrow {
	return New(time.Now())
}

func Yesterday() Arrow {
	return Now().Yesterday()
}

func Tomorrow() Arrow {
	return Now().Tomorrow()
}

func NextSecond() Arrow {
	return Now().AddSeconds(1).AtBeginningOfSecond()
}

func NextMinute() Arrow {
	return Now().AddMinutes(1).AtBeginningOfMinute()
}

func NextHour() Arrow {
	return Now().AddHours(1).AtBeginningOfHour()
}

func NextDay() Arrow {
	return Now().AddDays(1).AtBeginningOfDay()
}

func SleepUntil(t Arrow) {
	time.Sleep(t.Sub(Now()))
}

// Get the current time in the given timezone.
// The timezone parameter should correspond to a file in the IANA Time Zone database,
// such as "America/New_York".  "UTC" and "Local" are also acceptable.  If the timezone
// given isn't valid, then no change to the timezone is made.
func InTimezone(timezone string) Arrow {
	return Now().InTimezone(timezone)
}

func (a Arrow) Before(b Arrow) bool {
	return a.Time.Before(b.Time)
}

func (a Arrow) After(b Arrow) bool {
	return a.Time.After(b.Time)
}

func (a Arrow) Equal(b Arrow) bool {
	return a.Time.Equal(b.Time)
}

// Return an array of Arrow's from this one up to the given one,
// by duration.  For instance, Now().UpTo(Tomorrow(), Hour)
// will return an array of Arrow's from now until tomorrow by
// hour (inclusive of a and b).
func (a Arrow) UpTo(b Arrow, by time.Duration) []Arrow {
	var result []Arrow
	if a.After(b) {
		a, b = b, a
	}
	for a.Before(b) || a.Equal(b) {
		result = append(result, a)
		a = a.Add(by)
	}
	return result
}

func (a Arrow) Yesterday() Arrow {
	return a.AddDays(-1)
}

func (a Arrow) Tomorrow() Arrow {
	return a.AddDays(1)
}

func (a Arrow) UTC() Arrow {
	return New(a.Time.UTC())
}

func (a Arrow) Sub(b Arrow) time.Duration {
	return a.Time.Sub(b.Time)
}

// Add any duration parseable by time.ParseDuration
func (a Arrow) AddDuration(duration string) Arrow {
	if pduration, err := time.ParseDuration(duration); err == nil {
		return a.Add(pduration)
	}
	return a
}

func (a Arrow) Add(d time.Duration) Arrow {
	return New(a.Time.Add(d))
}

// The timezone parameter should correspond to a file in the IANA Time Zone database,
// such as "America/New_York".  "UTC" and "Local" are also acceptable.  If the timezone
// given isn't valid, then no change to the timezone is made.
func (a Arrow) InTimezone(timezone string) Arrow {
	if location, err := time.LoadLocation(timezone); err == nil {
		return New(a.In(location))
	}
	return a
}

func (a Arrow) AddDays(days int) Arrow {
	return New(a.AddDate(0, 0, days))
}

func (a Arrow) AddHours(hours int) Arrow {
	year, month, day := a.Time.Date()
	hour, min, sec := a.Time.Clock()
	d := time.Date(year, month, day, hour+hours, min, sec, a.Nanosecond(), a.Location())
	return New(d)
}

func (a Arrow) AddMinutes(minutes int) Arrow {
	year, month, day := a.Time.Date()
	hour, min, sec := a.Time.Clock()
	d := time.Date(year, month, day, hour, min+minutes, sec, a.Nanosecond(), a.Location())
	return New(d)
}

func (a Arrow) AddSeconds(seconds int) Arrow {
	year, month, day := a.Time.Date()
	hour, min, sec := a.Time.Clock()
	d := time.Date(year, month, day, hour, min, sec+seconds, a.Nanosecond(), a.Location())
	return New(d)
}

func (a Arrow) AtBeginningOfSecond() Arrow {
	return New(a.Truncate(Second))
}

func (a Arrow) AtBeginningOfMinute() Arrow {
	return New(a.Truncate(Minute))
}

func (a Arrow) AtBeginningOfHour() Arrow {
	return New(a.Truncate(Hour))
}

func (a Arrow) AtBeginningOfDay() Arrow {
	d := time.Duration(-a.Hour()) * Hour
	return a.AtBeginningOfHour().Add(d)
}

func (a Arrow) AtBeginningOfWeek() Arrow {
	days := time.Duration(-1*int(a.Weekday())) * Day
	return a.AtBeginningOfDay().Add(days)
}

func (a Arrow) AtBeginningOfMonth() Arrow {
	days := time.Duration(-1*int(a.Day())+1) * Day
	return a.AtBeginningOfDay().Add(days)
}

func (a Arrow) AtBeginningOfYear() Arrow {
	days := time.Duration(-1*int(a.YearDay())+1) * Day
	return a.AtBeginningOfDay().Add(days)
}

// Add any durations parseable by time.ParseDuration
func (a Arrow) AddDurations(durations ...string) Arrow {
	for _, duration := range durations {
		a = a.AddDuration(duration)
	}
	return a
}

func formatConvert(format string) string {
	// create mapping from strftime to time in Go
	strftimeMapping := map[string]string{
		"%a": "Mon",
		"%A": "Monday",
		"%b": "Jan",
		"%B": "January",
		"%c": "", // locale not supported
		"%C": "06",
		"%d": "02",
		"%D": "01/02/06",
		"%e": "_2",
		"%E": "", // modifiers not supported
		"%F": "2006-01-02",
		"%G": "%G", // special case, see below
		"%g": "%g", // special case, see below
		"%h": "Jan",
		"%H": "15",
		"%I": "03",
		"%j": "%j", // special case, see below
		"%k": "%k", // special case, see below
		"%l": "_3",
		"%m": "01",
		"%M": "04",
		"%n": "\n",
		"%O": "", // modifiers not supported
		"%p": "PM",
		"%P": "pm",
		"%r": "03:04:05 PM",
		"%R": "15:04",
		"%s": "%s", // special case, see below
		"%S": "05",
		"%t": "\t",
		"%T": "15:04:05",
		"%u": "%u", // special case, see below
		"%U": "%U", // special case, see below
		"%V": "%V", // special case, see below
		"%w": "%w", // special case, see below
		"%W": "%W", // special case, see below
		"%x": "%x", // locale not supported
		"%X": "%X", // locale not supported
		"%y": "06",
		"%Y": "2006",
		"%z": "-0700",
		"%Z": "MST",
		"%+": "Mon Jan _2 15:04:05 MST 2006",
		"%%": "%%", // special case, see below
	}

	for fmt, conv := range strftimeMapping {
		format = strings.Replace(format, fmt, conv, -1)
	}

	return format
}

// Parse the time using the same format string types as strftime
// See http://man7.org/linux/man-pages/man3/strftime.3.html for more info.
func CParse(layout, value string) (Arrow, error) {
	t, e := time.Parse(formatConvert(layout), value)
	return New(t), e
}

// Parse the time using the same format string types as strftime,
// within the given location.
// See http://man7.org/linux/man-pages/man3/strftime.3.html for more info.
func CParseInLocation(layout, value string, loc *time.Location) (Arrow, error) {
	t, e := time.ParseInLocation(formatConvert(layout), value, loc)
	return New(t), e
}

// Parse the time using the same format string types as strftime,
// within the given location (string value for timezone).
// See http://man7.org/linux/man-pages/man3/strftime.3.html for more info.
func CParseInStringLocation(layout, value, timezone string) (Arrow, error) {
	if location, err := time.LoadLocation(timezone); err == nil {
		return CParseInLocation(layout, value, location)
	} else {
		return New(time.Time{}), err
	}
}

// Format the time using the same format string types as strftime.
// See http://man7.org/linux/man-pages/man3/strftime.3.html for more info.
func (a Arrow) CFormat(format string) string {
	format = a.Format(formatConvert(format))

	year, week := a.ISOWeek()
	yearday := a.YearDay()
	weekday := a.Weekday()
	syear := strconv.Itoa(year)
	sweek := strconv.Itoa(week)
	syearday := strconv.Itoa(yearday)
	sweekday := strconv.Itoa(int(weekday))

	if a.Year() > 999 {
		format = strings.Replace(format, "%G", syear, -1)
		format = strings.Replace(format, "%g", syear[2:4], -1)
	}

	format = strings.Replace(format, "%j", syearday, -1)
	if a.Hour() < 10 {
		shour := " " + strconv.Itoa(a.Hour())
		format = strings.Replace(format, "%k", shour, -1)
	}
	format = strings.Replace(format, "%s", strconv.FormatInt(a.Unix(), 10), -1)

	if weekday == 0 {
		format = strings.Replace(format, "%u", "7", -1)
	} else {
		format = strings.Replace(format, "%u", sweekday, -1)
	}

	format = strings.Replace(format, "%U", weekNumber(a, time.Sunday), -1)
	format = strings.Replace(format, "%U", sweek, -1)
	format = strings.Replace(format, "%w", sweekday, -1)
	format = strings.Replace(format, "%W", weekNumber(a, time.Monday), -1)
	return strings.Replace(format, "%%", "%", -1)
}

// Used for %U and %W:
// %U: The week number of the current year as a decimal number, range
// 00 to 53, starting with the first Sunday as the first day of week 01.
//
// %W: The week number of the current year as a decimal number, range
// 00 to 53, starting with the first Monday as the first day of week 01.
func weekNumber(a Arrow, firstday time.Weekday) string {
	dayone := a.AtBeginningOfYear()
	for dayone.Weekday() != time.Sunday {
		dayone = dayone.AddDays(1)
	}
	week := int(a.Sub(dayone.AddDays(-7)) / Week)
	return strconv.Itoa(week)
}
