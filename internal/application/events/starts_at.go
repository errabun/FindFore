package events

import (
	"fmt"
	"strings"
	"time"

	"github.com/ericrabun/findfore-go/internal/domain/entity"
)

// ComposeStartsAt interprets date (YYYY-MM-DD) and teeTime (HH:MM or HH:MM:SS)
// as wall-clock time in the given IANA timezone (or DefaultCourseTimezone).
func ComposeStartsAt(date, teeTime, timezone string) (time.Time, error) {
	date = strings.TrimSpace(date)
	teeTime = strings.TrimSpace(teeTime)
	if date == "" || teeTime == "" {
		return time.Time{}, fmt.Errorf("date and tee_time are required")
	}

	loc, err := loadLocation(timezone)
	if err != nil {
		return time.Time{}, err
	}

	raw := date + " " + teeTime
	for _, layout := range []string{"2006-01-02 15:04", "2006-01-02 15:04:05"} {
		t, err := time.ParseInLocation(layout, raw, loc)
		if err == nil {
			return t, nil
		}
	}
	return time.Time{}, fmt.Errorf("invalid date/tee_time %q %q", date, teeTime)
}

// SplitStartsAt returns wall-clock date (YYYY-MM-DD) and tee time (HH:MM) in timezone.
func SplitStartsAt(startsAt time.Time, timezone string) (date, teeTime string, err error) {
	loc, err := loadLocation(timezone)
	if err != nil {
		return "", "", err
	}
	local := startsAt.In(loc)
	return local.Format("2006-01-02"), local.Format("15:04"), nil
}

func loadLocation(timezone string) (*time.Location, error) {
	timezone = strings.TrimSpace(timezone)
	if timezone == "" {
		timezone = entity.DefaultCourseTimezone
	}
	loc, err := time.LoadLocation(timezone)
	if err != nil {
		return nil, fmt.Errorf("invalid timezone %q: %w", timezone, err)
	}
	return loc, nil
}
