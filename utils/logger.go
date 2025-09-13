package utils

import (
	"os"

	"github.com/sirupsen/logrus"
)

var Log = logrus.New()

func InitLogger() {
	// output to stdout and file
	Log.Out = os.Stdout

	// log level
	Log.SetLevel(logrus.InfoLevel)

	// log format (JSON is convenient for ELK)
	Log.SetFormatter(&logrus.JSONFormatter{})

	if _, err := os.Stat("logs"); os.IsNotExist(err) {
    os.MkdirAll("logs", 0755)
}
	// you can add a file for logs
	file, err := os.OpenFile("logs/app.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err == nil {
		Log.SetOutput(file)
	} else {
		Log.Warn("Failed to open log file, outputting to stdout only")
	}
}
