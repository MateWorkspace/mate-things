package applicationnodemessagingcallback

import (
	"regexp"
	"strconv"
	"time"

	domainmodels "github.com/MateWorkspace/mate-things/backend/internal/domain/models"
)

var logLineRe = regexp.MustCompile(
	`^(\d{2})/(\d{2})/(\d{4}) (\d{2}):(\d{2}):(\d{2})\.(\d{3}) \[(\w+)\] \[([^\]]*)\] (.*)$`,
)

func parseLogLine(raw string) (level domainmodels.NodeLogLevel, tag string, message string, loggedAt time.Time) {
	match := logLineRe.FindStringSubmatch(raw)
	if match == nil {
		return invalidLogLine(raw)
	}

	day, dayErr := strconv.Atoi(match[1])
	month, monthErr := strconv.Atoi(match[2])
	year, yearErr := strconv.Atoi(match[3])
	hour, hourErr := strconv.Atoi(match[4])
	minute, minuteErr := strconv.Atoi(match[5])
	second, secondErr := strconv.Atoi(match[6])
	millisecond, millisecondErr := strconv.Atoi(match[7])
	level = parseLevel(match[8])
	if dayErr != nil || monthErr != nil || yearErr != nil || hourErr != nil || minuteErr != nil || secondErr != nil || millisecondErr != nil || level == domainmodels.NodeLogLevelNone && match[8] != string(domainmodels.NodeLogLevelNone) {
		return invalidLogLine(raw)
	}

	loggedAt = time.Date(year, time.Month(month), day, hour, minute, second, millisecond*int(time.Millisecond), time.UTC)
	if loggedAt.Year() != year || int(loggedAt.Month()) != month || loggedAt.Day() != day || loggedAt.Hour() != hour || loggedAt.Minute() != minute || loggedAt.Second() != second || loggedAt.Nanosecond() != millisecond*int(time.Millisecond) {
		return invalidLogLine(raw)
	}

	return level, match[9], match[10], loggedAt
}

func invalidLogLine(raw string) (domainmodels.NodeLogLevel, string, string, time.Time) {
	return domainmodels.NodeLogLevelNone, "", raw, time.Now().UTC()
}

func parseLevel(raw string) domainmodels.NodeLogLevel {
	switch domainmodels.NodeLogLevel(raw) {
	case domainmodels.NodeLogLevelNone, domainmodels.NodeLogLevelError, domainmodels.NodeLogLevelWarn, domainmodels.NodeLogLevelInfo, domainmodels.NodeLogLevelDebug:
		return domainmodels.NodeLogLevel(raw)
	default:
		return domainmodels.NodeLogLevelNone
	}
}
