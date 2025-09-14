package trace

var (
	severityNames = []string{
		DEBUG:  "debug",
		INFO:   "info",
		NOTICE: "notice",
		WARN:   "warn",
		ERR:    "err",
		CRIT:   "crit",
		ALERT:  "alert",
		EMERG:  "emerg",
	}

	severityNameMappingTable = map[string]Severity{
		"debug":  DEBUG,
		"info":   INFO,
		"notice": NOTICE,
		"warn":   WARN,
		"err":    ERR,
		"crit":   CRIT,
		"alert":  ALERT,
		"emerg":  EMERG,
	}
)

// Severity
type Severity int8

func (s Severity) Name() string {
	return severityNames[s]
}
