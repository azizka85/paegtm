package helpers

import "time"

func DateFormat(date interface{}) string {
	if date != nil {
		dt, ok := date.(time.Time)

		if ok {
			return dt.Format("2006-01-02")
		}
	}

	return ""
}

func DateMonthYearFormat(date interface{}) string {
	if date != nil {
		dt, ok := date.(time.Time)

		if ok {
			return dt.Format("2006-01")
		}
	}

	return ""
}

func DateEq(date1 string, date2 string) bool {
	return date1 == date2
}
