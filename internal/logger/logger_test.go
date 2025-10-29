package logger

import (
	"os"
	"testing"

	"github.com/sirupsen/logrus"
)

func TestInitLogger(t *testing.T) {
	t.Run("valid level", func(t *testing.T) {
		logger := InitLogger("debug")
		if logger.GetLevel() != logrus.DebugLevel {
			t.Errorf("expected log level %s, but got %s", logrus.DebugLevel, logger.GetLevel())
		}
	})

	t.Run("invalid level", func(t *testing.T) {
		logger := InitLogger("invalid")
		if logger.GetLevel() != logrus.InfoLevel {
			t.Errorf("expected log level %s, but got %s", logrus.InfoLevel, logger.GetLevel())
		}
	})
}

func TestSetLogFile(t *testing.T) {
	logger := InitLogger("info")

	t.Run("empty log file", func(t *testing.T) {
		SetLogFile(logger, "")
		if logger.Out == os.Stdout {
			t.Error("expected logger output to be a file")
		}
	})

	t.Run("custom log file", func(t *testing.T) {
		logFile := "test.log"
		SetLogFile(logger, logFile)
		if logger.Out == os.Stdout {
			t.Error("expected logger output to be a file")
		}
		os.Remove(logFile)
	})
}
