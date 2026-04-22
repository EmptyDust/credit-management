package timeutil

import "time"

const DateLayout = "2006-01-02"

var shanghai = loadShanghai()

func Shanghai() *time.Location {
	return shanghai
}

func Now() time.Time {
	return time.Now().In(Shanghai())
}

func ParseDate(value string) (time.Time, error) {
	return time.ParseInLocation(DateLayout, value, Shanghai())
}

func FormatDate(value time.Time) string {
	return value.In(Shanghai()).Format(DateLayout)
}

func NormalizeDate(value time.Time) time.Time {
	local := value.In(Shanghai())
	year, month, day := local.Date()
	return time.Date(year, month, day, 0, 0, 0, 0, Shanghai())
}

func loadShanghai() *time.Location {
	location, err := time.LoadLocation("Asia/Shanghai")
	if err == nil {
		return location
	}

	return time.FixedZone("CST", 8*60*60)
}
