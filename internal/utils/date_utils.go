package utils

import "time"

// DateAdjustForUtc returns converts time from 0 hour to 23
func DateAdjustForUtc(t time.Time) time.Time {
	return t.Add(time.Hour * 23)
}

// DateEqual returns true if the two dates are equal (check only year, month and day)
func DateEqual(date1, date2 time.Time) bool {
	return date1.Format("20060102") == date2.Format("20060102")
}

// DateFromString parses the string and returns time.
func DateFromString(sDate string) time.Time {
	t, err := time.Parse("2006-01-02", sDate)
	if err != nil {
		t, err = time.Parse("2006-01-2", sDate)
		if err != nil {
			t, err = time.Parse("2006-1-2", sDate)
			if err != nil {
				t, err = time.Parse("02/01/2006", sDate)
			}
		}
		// log.Printf("%s - Error: %v", sDate, err)
	}
	return t
}

// DateFormat1 formats a date to 2020-01-01
func DateFormat1(date time.Time) string {
	return date.Format("2006-01-02")
}

// DateForPeriod returns the date for the period
func DateForPeriod(period string) *time.Time {
	switch period {
	case "1W":
		return setDate(7)
	case "1M":
		return setDate(30)
	case "3M":
		return setDate(90)
	case "6M":
		return setDate(180)
	case "YTD":
		return setYtdDate()
	case "1Y":
		return setYearDate(-1)
	case "2Y":
		return setYearDate(-2)
	case "3Y":
		return setYearDate(-3)
	case "5Y":
		return setYearDate(-5)
	}
	return nil
}

func setDate(days int) *time.Time {

	var date time.Time
	date = time.Now().Add(-time.Hour * time.Duration(days*24))
	for IsWeekend(date) {
		date = date.Add(-time.Hour * time.Duration(24))
	}
	return &date
}

func setYtdDate() *time.Time {

	today := time.Now()
	date := time.Date(today.Year(), time.Month(1), 2, 0, 0, 0, 0, today.Location())

	for IsWeekend(date) {
		date = date.Add(time.Hour * time.Duration(24))
	}
	return &date
}

func setYearDate(offset int) *time.Time {
	today := time.Now()
	date := today.AddDate(offset, 0, 0)
	return &date
}

// IsWeekend returns true if the day is Saturday or Sunday
func IsWeekend(t time.Time) bool {
	return t.Weekday() == time.Saturday || t.Weekday() == time.Sunday
}
