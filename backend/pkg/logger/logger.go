package logger

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/sirupsen/logrus"
)

// CustomFormatter adalah custom formatter untuk logrus
type CustomFormatter struct{}

// Format mengimplementasikan logrus.Formatter interface
func (f *CustomFormatter) Format(entry *logrus.Entry) ([]byte, error) {
	logData := map[string]interface{}{
		"level": entry.Level.String(),
		"time":  entry.Time.Format(time.RFC3339),
	}

	// Jika ada fields, serialize sebagai JSON di msg field
	if len(entry.Data) > 0 {
		msgData := make(map[string]interface{}, len(entry.Data))
		for k, v := range entry.Data {
			// stringify error values — error objects marshal as {} (unexported fields)
			if err, ok := v.(error); ok {
				msgData[k] = err.Error()
			} else {
				msgData[k] = v
			}
		}
		msgBytes, _ := json.Marshal(msgData)
		logData["msg"] = string(msgBytes)
	} else {
		logData["msg"] = entry.Message
	}

	output, _ := json.Marshal(logData)
	output = append(output, '\n')
	return output, nil
}

func New(env string) *logrus.Logger {
	log := logrus.New()

	log.SetFormatter(&CustomFormatter{})

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
