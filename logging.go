package exoip

import (
	"fmt"
	"log"
	"log/syslog"
	"os"
)

type wrappedLogger struct {
	syslog       bool
	syslogWriter *syslog.Writer
	stdWriter    *log.Logger
}

// Logger represents a wrapped version of syslog
var Logger *wrappedLogger

// Warning logs a message with severity LOG_WARNING
func (l *wrappedLogger) Warning(msg string) {
	if l.syslog {
		l.syslogWriter.Warning(msg)
	} else {
		l.stdWriter.Printf("[WARNING] %s", msg)
	}
}

// Crit logs a message with severity LOG_CRIT
func (l *wrappedLogger) Crit(msg string) {
	if l.syslog {
		l.syslogWriter.Crit(msg)
	} else {
		l.stdWriter.Printf("[CRIT   ] %s", msg)
	}
}

// Info logs a message with severity LOG_INFO
func (l *wrappedLogger) Info(msg string) {
	if l.syslog {
		l.syslogWriter.Info(msg)
	} else {
		l.stdWriter.Printf("[INFO   ] %s", msg)
	}
}

func setupLogger(logStdout bool) {
	if logStdout {
		logger := log.New(os.Stdout, "exoip ", 0)
		Logger = &wrappedLogger{syslog: false, stdWriter: logger}
	} else {
		logger, err := syslog.New(syslog.LOG_DAEMON, "exoip")
		if err != nil {
			fmt.Fprintln(os.Stderr, "fatal error:", err)
			os.Exit(1)
		}
		Logger = &wrappedLogger{syslog: true, syslogWriter: logger}
	}
}
