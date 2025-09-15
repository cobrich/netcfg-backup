package server

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/cobrich/netcfg-backup/models"
	"github.com/gorilla/mux"
)

func (s *Server) handleDevicesList() http.HandlerFunc {
	type PageData struct {
		Devices []models.Device
	}

	return func(w http.ResponseWriter, r *http.Request) {
		devices, err := s.store.GetAllDevices()
		if err != nil {
			http.Error(w, "Failed to get devices", http.StatusInternalServerError)
			return
		}

		data := PageData{
			Devices: devices,
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
