package logger

import (
	"net/http"
	"os"
	"runtime"

	"github.com/sirupsen/logrus"
)

func InitLogger(logLevel string) *logrus.Logger {
	logger := logrus.New()
	logger.SetFormatter(&logrus.TextFormatter{
		FullTimestamp: true,
	})

	level, err := logrus.ParseLevel(logLevel)
	if err != nil {
		// Si le niveau de journalisation n'est pas reconnu, réglez par défaut sur InfoLevel
		logger.Warnf("Invalid log level '%s', falling back to 'info'", logLevel)
		level = logrus.InfoLevel
	}
	logger.SetLevel(level)

	return logger
}

func SetLogFile(logger *logrus.Logger, logFile string) {
	if logFile == "" {
		if runtime.GOOS == "linux" {
			logFile = "/var/log/jelly-wol-proxy.log"
		} else {
			logFile = "jelly-wol-proxy.log"
		}
	}

	file, err := os.OpenFile(logFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		logger.Warnf("Failed to open log file: %v", err)
	} else {
		logger.SetOutput(file)
	}
}

func LogRequest(logger *logrus.Logger, r *http.Request) {
	logger.WithFields(logrus.Fields{
		"client":     r.Header.Get("X-Forwarded-For"), // Replace the value of client by X-Forwarded-For
		"method":     r.Method,
		"user-agent": r.UserAgent(),
		"path":       r.URL.Path,
	}).Info()
}
