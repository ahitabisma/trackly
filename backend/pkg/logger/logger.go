package logger

import (
	"fmt"
	"io"
	"os"
	"time"

	"github.com/sirupsen/logrus"
)

func New(env string) *logrus.Logger {
	log := logrus.New()

	log.SetFormatter(&logrus.JSONFormatter{})

	date := time.Now().Format("2006-01-02")

	logDir := "storage/logs"
	os.MkdirAll(logDir, os.ModePerm)

	filePath := fmt.Sprintf("%s/%s.log", logDir, date)

	file, err := os.OpenFile(filePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		log.Fatal("Failed to open log file:", err)
	}

	if env == "production" {
		log.SetOutput(file)
		log.SetLevel(logrus.WarnLevel)
	} else {
		mw := io.MultiWriter(os.Stdout, file)
		log.SetOutput(mw)
		log.SetLevel(logrus.DebugLevel)
	}

	return log
}
