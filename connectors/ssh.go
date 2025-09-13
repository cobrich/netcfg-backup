package connectors

import (
	"context"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/cobrich/netcfg-backup/models"
	"github.com/cobrich/netcfg-backup/utils"

	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/knownhosts"
)

type SSHConnector struct {
	Host     string
	Username string
	Password string
	KeyPath  string
	Timeout  time.Duration
}

// Она создает метод аутентификации: либо по ключу, либо по паролю.
func createAuthMethod(keyPath, password string) (ssh.AuthMethod, error) {
	if keyPath != "" {
		key, err := os.ReadFile(keyPath)
		if err != nil {
			return nil, fmt.Errorf("не удалось прочитать приватный ключ из %s: %w", keyPath, err)
		}
		// Создаем signer из нашего ключа
		signer, err := ssh.ParsePrivateKey(key)
		if err != nil {
			// Если ключ зашифрован паролем, здесь будет ошибка.
			// Для простоты пока поддерживаем только незашифрованные ключи.
			return nil, fmt.Errorf("не удалось распарсить приватный ключ: %w", err)
		}
		return ssh.PublicKeys(signer), nil
	}
	// Если путь к ключу не указан, используем пароль
	return ssh.Password(password), nil
}

func (s *SSHConnector) RunCommands(cmds []string) ([]models.Result, error) {
	logger := utils.Log.WithField("host", s.Host)
	logger.Infof("SSH: подключение к %s...", s.Host)

	ctx, cancel := context.WithTimeout(context.Background(), s.Timeout)
	defer cancel()

	d := net.Dialer{}
	addr := s.Host
	if !strings.Contains(addr, ":") {
		addr = addr + ":22"
	}
	conn, err := d.DialContext(ctx, "tcp", addr)
	if err != nil {
		logger.Errorf("SSH: не удалось подключиться: %v", err)
		return nil, fmt.Errorf("не удалось подключиться к %s: %v", addr, err)
	}
	defer conn.Close()
	logger.Infof("SSH: соединение с %s установлено", addr)

	hostKeyCallback, err := createHostKeyCallback()
	if err != nil {
		return nil, fmt.Errorf("не удалось создать HostKeyCallback: %v", err)
	}

	authMethod, err := createAuthMethod(s.KeyPath, s.Password) // <-- s.KeyPath нужно добавить в структуру
	if err != nil {
		return nil, err
	}

	config := &ssh.ClientConfig{
		User:            s.Username,
		Auth:            []ssh.AuthMethod{authMethod}, // <-- Используем созданный метод
		HostKeyCallback: hostKeyCallback,
		Timeout:         s.Timeout,
	}

	c, chans, reqs, err := ssh.NewClientConn(conn, addr, config)
	if err != nil {
		logger.Errorf("SSH: не удалось создать SSH-сессию: %v", err)
		return nil, fmt.Errorf("не удалось установить SSH-сессию: %v", err)
	}
	client := ssh.NewClient(c, chans, reqs)
	defer client.Close()

	results := []models.Result{}

	for _, cmd := range cmds {
		logger.Infof("SSH: выполняем команду: %s", cmd)

		outputCh := make(chan []byte)
		errCh := make(chan error)

		go func(c string) {
			session, err := client.NewSession()
			if err != nil {
				errCh <- fmt.Errorf("не удалось создать сессию: %v", err)
				return
			}
			defer session.Close()

			output, err := session.CombinedOutput(c)
			if err != nil {
				errCh <- err
				return
			}
			outputCh <- output
		}(cmd)

		select {
		case <-ctx.Done():
			logger.Errorf("SSH: таймаут выполнения команды '%s'", cmd)
			results = append(results, models.Result{Cmd: cmd, Output: "таймаут"})
		case err := <-errCh:
			logger.Errorf("SSH: ошибка выполнения команды '%s': %v", cmd, err)
			results = append(results, models.Result{Cmd: cmd, Output: fmt.Sprintf("ошибка при выполнении: %v", err)})
		case output := <-outputCh:
			logger.Infof("SSH: команда '%s' успешно выполнена", cmd)
			results = append(results, models.Result{Cmd: cmd, Output: string(output)})
		}
	}

	return results, nil
}

// Новая вспомогательная функция для создания HostKeyCallback
func createHostKeyCallback() (ssh.HostKeyCallback, error) {
	// Находим домашнюю директорию текущего пользователя
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("не удалось найти домашнюю директорию: %v", err)
	}

	// Формируем путь к файлу known_hosts
	knownHostsPath := filepath.Join(homeDir, ".ssh", "known_hosts")

	// Создаем callback. knownhosts.New автоматически обработает создание файла, если его нет.
	callback, err := knownhosts.New(knownHostsPath)
	if err != nil {
		return nil, fmt.Errorf("не удалось прочитать файл known_hosts (%s): %v", knownHostsPath, err)
	}

	return callback, nil
}
