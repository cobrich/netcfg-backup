package utils

import (
	"fmt"
	"os"
	"path/filepath" // Очень важно для безопасного соединения путей
	"ssh-fetcher/models"
	"time"
)

// CreateBackup теперь принимает базовый путь
func CreateBackup(basePath string) error {
	if _, err := os.Stat(basePath); os.IsNotExist(err) {
		// Используем MkdirAll, так как он не выдает ошибку, если папка уже существует
		err = os.MkdirAll(basePath, 0755)
		if err != nil {
			Log.Fatalf("❌ Ошибка создания папки %s: %v", basePath, err)
			return err
		}
	}
	return nil
}

// GetDeviceDir теперь принимает базовый путь
func GetDeviceDir(basePath, host string) (string, error) {
	// Безопасно соединяем пути: basePath + "/" + host
	dir := filepath.Join(basePath, host)
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		if err := os.MkdirAll(dir, 0755); err != nil {
			Log.Errorf("не удалось создать папку %s: %v", dir, err)
			return "", err
		}
	}
	return dir, nil
}

// GetBackupFilename теперь принимает базовый путь
func GetBackupFilename(basePath, host string) (string, error) {
	// Сначала получаем директорию для устройства, передавая basePath
	dir, err := GetDeviceDir(basePath, host)
	if err != nil {
		Log.Error(err)
		return "", err
	}
	tstamp := time.Now().Format("2006-01-02_15-04-05")
	filename := fmt.Sprintf("backup_%s.txt", tstamp)
	
	// Безопасно соединяем путь к директории и имя файла
	return filepath.Join(dir, filename), nil
}

// WriteResultsToFile теперь принимает базовый путь
func WriteResultsToFile(basePath string, dev models.Device, results []models.Result) error {
	// Получаем полное имя файла, передавая basePath
	filename, err := GetBackupFilename(basePath, dev.Host)
	entry := Log.WithFields(map[string]interface{}{
		"host": dev.Host,
	})
	if err != nil {
		entry.Error(err)
		return err
	}

	f, err := os.Create(filename)
	if err != nil {
		entry.WithField("file", filename).Errorf("Не удалось создать файл: %v", err)
		return err
	}
	defer f.Close()

	// шапка файла
	f.WriteString("########################################\n")
	f.WriteString(fmt.Sprintf(" Host: %s\n", dev.Host))
	f.WriteString(fmt.Sprintf(" User: %s\n", dev.Username))
	f.WriteString(fmt.Sprintf(" Date: %s\n", time.Now().Format("2006-01-02 15:04:05")))
	f.WriteString("########################################\n\n")

	// вывод команд
	for _, result := range results {
		f.WriteString("### " + result.Cmd + " ###\n")
		f.WriteString(result.Output + "\n\n")
	}

	entry.WithField("file", filename).Info("✅ Результат сохранён")
	return nil
}