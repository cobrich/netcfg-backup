package main

import (
	"flag"
	"log"
	"os"
	"ssh-fetcher/config"
	"ssh-fetcher/connectors"
	"ssh-fetcher/models"
	"ssh-fetcher/utils"
	"sync"
	"time"

	"github.com/joho/godotenv"
)

const defaultTimeout = 10 * time.Second
const numWorkers = 10

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Println("Warning: .env file not found, using system environment variables")
	}

	backupPath := flag.String("backup-path", "backups", "Путь к папке для сохранения бэкапов")
	flag.Parse()

	utils.InitLogger()

	err = utils.CreateBackup(*backupPath)
	if err != nil {
		utils.Log.WithField("component", "backup").Error(err)
		return
	}

	devices := config.ReadConfig()

	utils.Log.Infof("Загружено %d устройств из конфигурации", len(devices))
	if len(devices) == 0 {
		utils.Log.Warn("Список устройств пуст. Завершение работы.")
		return
	}

	jobs := make(chan models.Device, len(devices))

	var wg sync.WaitGroup

	for w := 1; w <= numWorkers; w++ {
		wg.Add(1)
		go worker(&wg, w, jobs, *backupPath)
	}

	for _, dev := range devices {
		jobs <- dev
	}
	close(jobs)

	wg.Wait()

	utils.Log.Info("Все задачи выполнены.")
}

func worker(wg *sync.WaitGroup, id int, jobs <-chan models.Device, backupPath string) {
	defer wg.Done()

	for dev := range jobs {
		entry := utils.Log.WithFields(map[string]interface{}{
			"worker_id": id,
			"host":      dev.Host,
			"protocol":  dev.Protocol,
		})
		entry.Info("Воркер взял задачу")

		if dev.PasswordEnv != "" {
			dev.Password = os.Getenv(dev.PasswordEnv)
			if dev.Password == "" {
				entry.Warnf("Переменная окружения '%s' не задана или пуста", dev.PasswordEnv)
			}
		}

		timeout := defaultTimeout
		if dev.TimeoutSeconds > 0 {
			timeout = time.Duration(dev.TimeoutSeconds) * time.Second
		}

		var connector connectors.Connector
		switch dev.Protocol {
		case "ssh":
			connector = &connectors.SSHConnector{
				Host:     dev.Host,
				Username: dev.Username,
				Password: dev.Password,
				KeyPath:  dev.KeyPath,
				Timeout:  timeout,
			}
		case "telnet":
			connector = &connectors.TelnetConnector{
				Host:     dev.Host,
				Username: dev.Username,
				Password: dev.Password,
				Prompt:   dev.Prompt,
				Timeout:  timeout,
			}
		default:
			entry.Error("Неизвестный протокол")
			continue
		}

		results, err := connector.RunCommands(dev.Commands)
		if err != nil {
			entry.WithField("error", err).Error("Ошибка выполнения команд")
			continue
		}

		err = utils.WriteResultsToFile(backupPath, dev, results)
		if err != nil {
			entry.WithField("error", err).Error("Ошибка сохранения результатов")
		} else {
			entry.Info("Результаты успешно сохранены")
		}
	}
}
