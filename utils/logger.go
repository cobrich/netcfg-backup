package utils

import (
	"os"

	"github.com/sirupsen/logrus"
)

var Log = logrus.New()

func InitLogger() {
	// вывод в stdout и в файл
	Log.Out = os.Stdout

	// уровень логов
	Log.SetLevel(logrus.InfoLevel)

	// формат логов (JSON удобен для ELK)
	Log.SetFormatter(&logrus.JSONFormatter{})

	if _, err := os.Stat("logs"); os.IsNotExist(err) {
    os.MkdirAll("logs", 0755)
}
	// можно добавить файл для логов
	file, err := os.OpenFile("logs/app.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err == nil {
		Log.SetOutput(file)
	} else {
		Log.Warn("Не удалось открыть файл логов, вывод только в stdout")
	}
}
