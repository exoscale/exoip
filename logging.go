package exoip

import (
	"fmt"
	"log"
	"log/syslog"
	"os"
)

// Warning logs a message with severity LOG_WARNING
func (l *wrappedLogger) Warning(msg string, v ...interface{}) {
	if l.syslog {
		err := l.syslogWriter.Warning(fmt.Sprintf(msg, v...))
		if err != nil {
			panic(err)
		}
	} else {
		l.stdWriter.Printf("[WARNING] "+msg, v...)
	}
}

// Crit logs a message with severity LOG_CRIT
func (l *wrappedLogger) Crit(msg string, v ...interface{}) {
	if l.syslog {
		err := l.syslogWriter.Crit(fmt.Sprintf(msg, v...))
		if err != nil {
			panic(err)
		}
	} else {
		l.stdWriter.Printf("[CRIT   ] "+msg, v...)
	}
}

// Info logs a message with severity LOG_INFO
func (l *wrappedLogger) Info(msg string, v ...interface{}) {
	if l.syslog {
		err := l.syslogWriter.Info(fmt.Sprintf(msg, v...))
		if err != nil {
			panic(err)
		}
	} else if Verbose {
		l.stdWriter.Printf("[INFO   ] "+msg, v...)
	}
}

// SetupLogger initializes the logger
func SetupLogger(logStdout bool) {
	if logStdout {
		logger := log.New(os.Stdout, "exoip ", 0)
		Logger = &wrappedLogger{syslog: false, stdWriter: logger}
	} else {
		logger, err := syslog.New(syslog.LOG_DAEMON, "exoip")
		if err != nil {
			_, errP := fmt.Fprintf(os.Stderr, "fatal error: %s\n", err)
			if errP != nil {
				panic(errP)
			}
			os.Exit(1)
		}
		Logger = &wrappedLogger{syslog: true, syslogWriter: logger}
	}
}
