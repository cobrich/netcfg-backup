package connectors

import (
	"fmt"
	"net" // ИЗМЕНЕНИЕ: нужен для DialTimeout
	"strings"
	"time"

	"github.com/cobrich/netcfg-backup/models"
	"github.com/cobrich/netcfg-backup/utils"

	"github.com/ziutek/telnet"
)

type TelnetConnector struct {
	Host     string
	Username string
	Password string
	Prompt   string
	Timeout  time.Duration // Этот таймаут теперь будет для каждой операции
}

// Устанавливаем разумные таймауты по умолчанию
const defaultTelnetTimeout = 15 * time.Second

func (t *TelnetConnector) RunCommands(cmds []string) ([]models.Result, error) {
	logger := utils.Log.WithField("host", t.Host)
	logger.Infof("Telnet: подключение к %s...", t.Host)

	// ИЗМЕНЕНИЕ: Используем net.DialTimeout для подключения с таймаутом
	addr := t.Host
	if !strings.Contains(addr, ":") {
		addr = addr + ":23"
	}
	connDialer, err := net.DialTimeout("tcp", addr, t.getTimeout())
	if err != nil {
		logger.Errorf("Telnet: не удалось подключиться: %v", err)
		return nil, fmt.Errorf("telnet: не удалось подключиться: %v", err)
	}

	// Оборачиваем соединение в telnet.Conn
	conn, err := telnet.NewConn(connDialer)
	if err != nil {
		// Эта ошибка может возникнуть, если Telnet handshake (обмен опциями) не удался
		logger.Errorf("Telnet: не удалось создать Telnet-сессию: %v", err)
		connDialer.Close() // Закрываем базовое TCP соединение
		return nil, fmt.Errorf("telnet: не удалось создать Telnet-сессию: %v", err)
	}

	defer conn.Close()

	// Устанавливаем дедлайн на все соединение. Он будет сдвигаться для каждой операции
	conn.SetUnixWriteMode(true)

	// --- Авторизация с дедлайнами ---
	if err := expect(conn, t.getTimeout(), "Username:", "login:"); err != nil {
		return nil, fmt.Errorf("telnet: не дождался приглашения для ввода имени пользователя: %v", err)
	}
	if err := send(conn, t.getTimeout(), t.Username); err != nil {
		return nil, fmt.Errorf("telnet: не удалось отправить имя пользователя: %v", err)
	}

	if err := expect(conn, t.getTimeout(), "Password:", "password:"); err != nil {
		return nil, fmt.Errorf("telnet: не дождался приглашения для ввода пароля: %v", err)
	}
	if err := send(conn, t.getTimeout(), t.Password); err != nil {
		return nil, fmt.Errorf("telnet: не удалось отправить пароль: %v", err)
	}

	if t.Prompt == "" {
		t.Prompt = ">"
	}
	// Ожидаем prompt после входа
	if _, err := readUntil(conn, t.getTimeout(), t.Prompt); err != nil {
		return nil, fmt.Errorf("telnet: не дождался prompt после входа: %v", err)
	}

	results := []models.Result{}

	for _, cmd := range cmds {
		logger.Infof("Telnet: выполняем команду: %s", cmd)

		if err := send(conn, t.getTimeout(), cmd); err != nil {
			results = append(results, models.Result{Cmd: cmd, Output: fmt.Sprintf("ошибка при отправке: %v", err)})
			continue // Переходим к следующей команде
		}

		output, err := readUntil(conn, t.getTimeout(), t.Prompt)
		if err != nil {
			logger.Errorf("Telnet: ошибка выполнения команды '%s': %v", cmd, err)
			results = append(results, models.Result{Cmd: cmd, Output: fmt.Sprintf("ошибка при выполнении: %v", err)})
		} else {
			// ИЗМЕНЕНИЕ: Очищаем вывод от эха команды и prompt
			cleanOutput := cleanTelnetOutput(output, cmd, t.Prompt)
			logger.Infof("Telnet: команда '%s' успешно выполнена", cmd)
			results = append(results, models.Result{Cmd: cmd, Output: cleanOutput})
		}
	}

	return results, nil
}

// ИЗМЕНЕНИЕ: Вспомогательная функция для получения таймаута с значением по умолчанию
func (t *TelnetConnector) getTimeout() time.Duration {
	if t.Timeout > 0 {
		return t.Timeout
	}
	return defaultTelnetTimeout
}

// ИЗМЕНЕНИЕ: Вспомогательная функция для ожидания одного из нескольких вариантов текста
func expect(conn *telnet.Conn, timeout time.Duration, delimiters ...string) error {
	conn.SetReadDeadline(time.Now().Add(timeout))
	return conn.SkipUntil(delimiters...)
}

// ИЗМЕНЕНИЕ: Вспомогательная функция для отправки данных с таймаутом
func send(conn *telnet.Conn, timeout time.Duration, s string) error {
	conn.SetWriteDeadline(time.Now().Add(timeout))
	_, err := conn.Write([]byte(s + "\n"))
	return err
}

// ИЗМЕНЕНИЕ: Улучшенная функция чтения до prompt с таймаутом
func readUntil(conn *telnet.Conn, timeout time.Duration, prompt string) (string, error) {
	conn.SetReadDeadline(time.Now().Add(timeout))
	data, err := conn.ReadUntil(prompt)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// ИЗМЕНЕНИЕ: Вспомогательная функция для очистки вывода
func cleanTelnetOutput(output, cmd, prompt string) string {
	// Удаляем эхо команды (если оно есть в начале)
	output = strings.TrimPrefix(output, cmd)
	// Удаляем prompt (если он есть в конце)
	output = strings.TrimSuffix(output, prompt)
	// Убираем лишние пробелы и переносы строк по краям
	return strings.TrimSpace(output)
}
