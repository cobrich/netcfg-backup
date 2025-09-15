package server

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/cobrich/netcfg-backup/backups"
	"github.com/cobrich/netcfg-backup/models"
	"github.com/cobrich/netcfg-backup/utils"
	"github.com/gorilla/mux"
)

func (s *Server) handleDevicesList() http.HandlerFunc {
	type PageData struct {
		Devices         []models.Device
		FlashMessages   []interface{}
		IsBackupRunning bool
	}

	return func(w http.ResponseWriter, r *http.Request) {
		devices, err := s.store.GetAllDevices()
		if err != nil {
			http.Error(w, "Failed to get devices", http.StatusInternalServerError)
			return
		}

		session, _ := s.sessionStore.Get(r, "netcfg-backup-session")
		flashes := session.Flashes()
		session.Save(r, w)

		data := PageData{
			Devices:         devices,
			FlashMessages:   flashes,
			IsBackupRunning: s.isBackupRunning,
		}

		renderTemplate(w, "devices.html", data)
	}
}

func (s *Server) handleDeviceAddForm() http.HandlerFunc {
	type PageData struct {
		Device      models.Device
		CommandsStr string
	}
	return func(w http.ResponseWriter, r *http.Request) {
		renderTemplate(w, "device_form.html", PageData{})
	}
}

func (s *Server) handleDeviceAddSubmit() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := r.ParseForm(); err != nil {
			http.Error(w, "Failed to parse form", http.StatusBadRequest)
			return
		}

		commandsStr := r.FormValue("commands")
		var commands []string
		for _, cmd := range strings.Split(commandsStr, "\n") {
			cmd = strings.TrimSpace(cmd)
			if cmd != "" {
				commands = append(commands, cmd)
			}
		}

		newDevice := models.Device{
			Host:        r.FormValue("host"),
			Username:    r.FormValue("username"),
			Protocol:    r.FormValue("protocol"),
			KeyPath:     r.FormValue("key_path"),
			PasswordEnv: r.FormValue("password_env"),
			Prompt:      r.FormValue("prompt"),
			Commands:    commands,
		}

		if err := s.store.AddDevice(newDevice); err != nil {
			http.Error(w, fmt.Sprintf("Failed to add device: %v", err), http.StatusInternalServerError)
			return
		}
		http.Redirect(w, r, "/", http.StatusSeeOther)
	}
}

func (s *Server) handleDeviceEditForm() http.HandlerFunc {
	type PageData struct {
		Device      models.Device
		CommandsStr string
	}
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		host := vars["host"]
		device, err := s.store.GetDeviceByHost(host)
		if err != nil {
			http.Error(w, "Device not found", http.StatusNotFound)
			return
		}

		commandsStr := strings.Join(device.Commands, "\n")

		renderTemplate(w, "device_form.html", PageData{Device: *device, CommandsStr: commandsStr})
	}
}

func (s *Server) handleDeviceEditSubmit() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		host := vars["host"]
		if err := r.ParseForm(); err != nil {
			http.Error(w, "Failed to parse form", http.StatusBadRequest)
			return
		}

		commandsStr := r.FormValue("commands")
		var commands []string
		for _, cmd := range strings.Split(commandsStr, "\n") {
			cmd = strings.TrimSpace(cmd)
			if cmd != "" {
				commands = append(commands, cmd)
			}
		}

		updatedDevice := models.Device{
			Host:        host,
			Username:    r.FormValue("username"),
			Protocol:    r.FormValue("protocol"),
			KeyPath:     r.FormValue("key_path"),
			PasswordEnv: r.FormValue("password_env"),
			Prompt:      r.FormValue("prompt"),
			Commands:    commands,
		}

		if err := s.store.UpdateDevice(updatedDevice); err != nil {
			http.Error(w, fmt.Sprintf("Failed to update device: %v", err), http.StatusInternalServerError)
			return
		}
		http.Redirect(w, r, "/", http.StatusSeeOther)
	}
}

// handleDeviceRemove deletes a device.
func (s *Server) handleDeviceRemove() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		host := vars["host"]

		if err := s.store.RemoveDevice(host); err != nil {
			http.Error(w, fmt.Sprintf("Failed to remove device: %v", err), http.StatusInternalServerError)
			return
		}

		http.Redirect(w, r, "/", http.StatusSeeOther)
	}
}

// handleBackupHostsList shows a list of hosts that have backups.
func (s *Server) handleBackupHostsList() http.HandlerFunc {
	type PageData struct {
		Hosts []string
	}
	return func(w http.ResponseWriter, r *http.Request) {
		hosts, err := s.backupService.ListBackedUpHosts()
		if err != nil {
			http.Error(w, "Failed to list backup hosts", http.StatusInternalServerError)
			return
		}
		renderTemplate(w, "backups_hosts.html", PageData{Hosts: hosts})
	}
}

// handleBackupFilesList shows a list of backup files for a specific host.
func (s *Server) handleBackupFilesList() http.HandlerFunc {
	type PageData struct {
		Host    string
		Backups []backups.BackupInfo
	}
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		host := vars["host"]
		backupList, err := s.backupService.ListBackupsForHost(host)
		if err != nil {
			http.Error(w, "Failed to list backups for host", http.StatusInternalServerError)
			return
		}
		renderTemplate(w, "backups_files.html", PageData{Host: host, Backups: backupList})
	}
}

// handleBackupView shows the content of a single backup file.
func (s *Server) handleBackupView() http.HandlerFunc {
	type PageData struct {
		Host     string
		Filename string
		Content  string
	}
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		host := vars["host"]
		filename := vars["filename"]
		content, err := s.backupService.GetBackupContent(host, filename)
		if err != nil {
			http.Error(w, "Failed to read backup file", http.StatusInternalServerError)
			return
		}
		renderTemplate(w, "backup_view.html", PageData{Host: host, Filename: filename, Content: content})
	}
}

// handleRunBackup triggers the backup process in the background.
func (s *Server) handleRunBackup() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if s.isBackupRunning {
			http.Error(w, "A backup process is already running.", http.StatusConflict)
			return
		}

		go func() {
			s.isBackupRunning = true
			defer func() { s.isBackupRunning = false }()

			if err := s.coreService.Run(); err != nil {
				utils.Log.Errorf("Background backup run failed: %v", err)
			}
		}()

		session, _ := s.sessionStore.Get(r, "netcfg-backup-session")
		session.AddFlash("✅ Backup process started in the background!")
		session.Save(r, w)

		http.Redirect(w, r, "/", http.StatusSeeOther)
	}
}
