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

func (s Severity) Value() int {
	if s.IsValid() {
		return int(s)
	}
	return __severity_none__
}

func (s Severity) IsValid() bool {
	if (s > __severity_maximum__) || (s < __severity_minimum__) {
		return false
	}
	return true
}

func (s Severity) Name() string {
	return severityNames[s]
}

func (s Severity) String() string {
	return s.Name()
}

func ParseSeverity(name string) Severity {
	v, ok := severityNameMappingTable[name]
	if ok {
		return v
	}
	return __severity_none__
}
