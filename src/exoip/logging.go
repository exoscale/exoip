package exoip

import (
	"log/syslog"
)

var Logger *syslog.Writer

func SetupLogger() {

	logger, err := syslog.New(syslog.LOG_DAEMON, "exoip")
	AssertSuccess(err)
	Logger = logger
}
