package utils

import (
	"regexp"
	"strconv"
	"time"
)

var TimeLocation *time.Location

func LoadLocation(name string) {
	TimeLocation, _ = time.LoadLocation(name)
}

func ParseDurationPattern(durationPattern string, before bool) (time.Time, bool) {
	multipliers := map[string]time.Duration{
		"SECONDS": time.Second,
		"SECOND":  time.Second,
		"MINUTES": time.Minute,
		"MINUTE":  time.Minute,
		"HOURS":   time.Hour,
		"HOUR":    time.Hour,
		"DAYS":    time.Hour * 24,
		"DAY":     time.Hour * 24,
		"WEEKS":   time.Hour * 24 * 7,
		"WEEK":    time.Hour * 24 * 7,
		"MONTHS":  time.Hour * 24 * 30,
		"MONTH":   time.Hour * 24 * 30,
		"YEARS":   time.Hour * 24 * 365,
	}
	regexPattern := `((\d+)\s(SECONDS|MINUTES|HOURS|DAYS|WEEKS|MONTHS|YEARS|SECOND|MINUTE|HOUR|DAY|WEEK|MONTH|YEAR))`
	matches := regexp.MustCompile(regexPattern).FindAllStringSubmatch(durationPattern, -1)
	t := time.Now().In(TimeLocation)
	success := false
	for _, match := range matches {
		parsedInt, _ := strconv.ParseInt(match[2], 10, 64)
		bfMul := time.Duration(1)
		if before {
			bfMul = -1
		}
		t = t.Add(bfMul * time.Duration(parsedInt) * multipliers[match[3]])
		success = true
	}
	return t, success
}
