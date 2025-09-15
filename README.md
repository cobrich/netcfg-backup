# Netcfg-Backup

A reliable, web-based tool for automated backup of network device configurations, featuring a full observability stack.

[![Go Report Card](https://goreportcard.com/badge/github.com/cobrich/netcfg-backup)](https://goreportcard.com/report/github.com/cobrich/netcfg-backup)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

`netcfg-backup` provides a user-friendly web interface to manage your network device inventory, run backups on demand, and browse historical configuration files. It uses a concurrent worker pool to handle many devices efficiently and comes with a pre-configured monitoring stack using Prometheus and Grafana.

For automation and scripting, a powerful command-line interface (CLI) is also available.

## Key Features

-   **Web Interface:** A clean, intuitive UI for all primary operations:
    -   Full CRUD (Create, Read, Update, Delete) for your device inventory.
    -   On-demand backup execution for all devices.
    -   A browser for viewing saved backup files.
-   **Persistent Storage:** Uses a local SQLite database to reliably store device configurations.
-   **Built-in Monitoring Stack:** Comes with a `docker-compose` setup for Prometheus and Grafana, providing instant insights into job performance and success rates.
-   **Versatile CLI:** A powerful command-line interface for scripting and automation (`add`, `list`, `edit`, `remove`, `run`, `exec`, `migrate`).
-   **Multi-protocol & Secure:** Connects via SSH (keys) or Telnet, handling secrets securely via environment variables.

## Getting Started

### Prerequisites

-   [Go](https://golang.org/doc/install) (version 1.24 or later)
-   [Docker](https://docs.docker.com/get-docker/) and Docker Compose (v2)
-   Git

### Installation & First Run

The recommended way to run `netcfg-backup` is using the full stack via Docker Compose.

1.  **Clone the repository:**
    ```bash
    git clone https://github.com/cobrich/netcfg-backup.git
    cd netcfg-backup
    ```

2.  **Set up secrets:**
    Copy the example `.env.example` to `.env` and define any necessary passwords. You will need to create a `SESSION_AUTH_KEY` for the web UI.
    ```bash
    cp .env.example .env
    # Now edit .env with your favorite editor
    ```

3.  **Launch the application stack:**
    ```bash
    docker compose up --build
    ```
    This command will build and start the `netcfg-backup` (in `server` mode), `prometheus`, and `grafana` containers.

4.  **Access the Web UI:**
    Open your browser and navigate to `http://localhost:8080`.
    
    You can now manage your devices, run backups, and view results directly from the web interface.

## Usage

### Web Interface

-   **Devices:** `http://localhost:8080/` — Main page for listing, adding, editing, and removing devices.
-   **Backups:** `http://localhost:8080/backups` — Browse backups by host and view their content.
-   **Run Backup:** The "Run Backup Now" button on the main page triggers the backup process for all configured devices in the background.

### Monitoring

-   **Prometheus:** `http://localhost:9091` — View raw metrics and target status.
-   **Grafana:** `http://localhost:3000` (login: admin/admin) — Build dashboards to visualize `netcfg_backup_*` metrics.

### Command-Line Interface (CLI)

The CLI is perfect for scripting, automation, or quick ad-hoc tasks.

1.  **Build the binary:**
    ```bash
    go build -o netcfg-backup .
    ```

2.  **Available Commands:**
    -   `./netcfg-backup server`: Starts the web server (this is what Docker Compose uses).
    -   `./netcfg-backup run`: Runs the backup process for all devices in the database.
    -   `./netcfg-backup list | add | edit | remove`: Manage the device inventory from the command line.
    -   `./netcfg-backup exec --host ...`: Execute ad-hoc commands on a single device.
    -   `./netcfg-backup migrate`: One-time command to migrate devices from an old `devices.json` file.

    For more details on any command, use the `--help` flag, e.g., `./netcfg-backup exec --help`.

## License

This project is licensed under the MIT License. See the `LICENSE` file for details.