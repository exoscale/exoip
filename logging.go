package exoip

import (
	"fmt"
	"log"
	"log/syslog"
	"os"
)

type WrappedLogger struct {
	syslog        bool
	syslog_writer *syslog.Writer
	std_writer    *log.Logger
}

var Logger *WrappedLogger

func (l *WrappedLogger) Warning(msg string) {
	if l.syslog {
		l.syslog_writer.Warning(msg)
	} else {
		l.std_writer.Printf("[WARNING] %s", msg)
	}
}

func (l *WrappedLogger) Crit(msg string) {
	if l.syslog {
		l.syslog_writer.Crit(msg)
	} else {
		l.std_writer.Printf("[CRIT   ] %s", msg)
	}
}

func (l *WrappedLogger) Info(msg string) {
	if l.syslog {
		l.syslog_writer.Info(msg)
	} else {
		l.std_writer.Printf("[INFO   ] %s", msg)
	}
}

func SetupLogger(log_stdout bool) {

	if log_stdout {
		logger := log.New(os.Stdout, "exoip ", 0)
		Logger = &WrappedLogger{syslog: false, std_writer: logger}
	} else {
		logger, err := syslog.New(syslog.LOG_DAEMON, "exoip")
		if err != nil {
			fmt.Fprintln(os.Stderr, "fatal error:", err)
			os.Exit(1)
		}
		Logger = &WrappedLogger{syslog: true, syslog_writer: logger}
	}
}
