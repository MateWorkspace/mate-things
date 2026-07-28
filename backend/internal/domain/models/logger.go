package domainmodels

type LoggerFormat string

const (
	LoggerFormatPlain LoggerFormat = "plain"
	LoggerFormatJson  LoggerFormat = "json"
)

type LoggerLevel string

const (
	LoggerLevelNone  LoggerLevel = "NONE "
	LoggerLevelError LoggerLevel = "ERROR"
	LoggerLevelWarn  LoggerLevel = "WARN "
	LoggerLevelInfo  LoggerLevel = "INFO "
	LoggerLevelDebug LoggerLevel = "DEBUG"
)

func (l LoggerLevel) String() string {
	return string(l)
}

func (l LoggerLevel) IsValid() bool {
	switch l {
	case LoggerLevelNone, LoggerLevelError, LoggerLevelWarn, LoggerLevelInfo, LoggerLevelDebug:
		return true
	default:
		return false
	}
}

func (l LoggerLevel) Order() int {
	switch l {
	case LoggerLevelNone:
		return 0
	case LoggerLevelError:
		return 1
	case LoggerLevelWarn:
		return 2
	case LoggerLevelInfo:
		return 3
	case LoggerLevelDebug:
		return 4
	default:
		return -1
	}
}

type LoggerMeta map[string]any
