# Netcfg-Backup

A reliable and extensible tool written in Go for automated backup of network device configurations, featuring a built-in monitoring stack.

[![Go Report Card](https://goreportcard.com/badge/github.com/cobrich/netcfg-backup)](https://goreportcard.com/report/github.com/cobrich/netcfg-backup)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

`netcfg-backup` allows you to connect to a large number of network devices concurrently via SSH or Telnet, execute a predefined set of commands, and save the output. It comes with a pre-configured observability stack using Prometheus and Grafana to monitor job performance.

## Key Features

-   **Built-in Monitoring Stack:** Comes with a ready-to-use `docker-compose` setup for Prometheus and Grafana, providing instant insights into job performance, duration, and success rates.
-   **Versatile CLI:** Manage your device inventory with `add`, `list`, `edit`, `remove`, run scheduled backups with `run`, or execute one-off commands with `exec`.
-   **Multi-protocol Support:** Connect to devices using SSH (with key-based authentication) or legacy Telnet.
-   **Concurrent Operations:** Efficiently handles large device lists using a worker pool.
-   **Secure Credential Management:** Handles secrets securely via environment variables, with support for `.env` files for easy local development.
-   **DevOps Ready:** Features structured JSON logging and a multi-stage `Dockerfile` for quick and secure containerization.

## Getting Started

### Prerequisites

-   [Go](https://golang.org/doc/install) (version 1.24 or later)
-   [Docker](https://docs.docker.com/get-docker/) and Docker Compose (v2)
-   Git

### Installation

1.  Clone the repository:
    ```bash
    git clone https://github.com/cobrich/netcfg-backup.git
    cd netcfg-backup
    ```

2.  **Configuration:**
    -   Copy `.env.example` to `.env` and fill in your secrets.
    -   To use the `run` command, add devices to your inventory first: `./netcfg-backup add`. This will create the `devices/devices.json` file.

## Usage

`netcfg-backup` is a command-line tool with several subcommands.

### Running the Full Monitoring Stack (Recommended)

This is the easiest way to run scheduled backups and see all features in action.
```bash
docker compose up --build
```
This command starts the `netcfg-backup` (in `run` mode), `prometheus`, and `grafana` containers.

-   **Prometheus:** `http://localhost:9091`
-   **Grafana:** `http://localhost:3000` (login: admin/admin)

### Standalone CLI Usage

You can also build and run the tool directly for quick tasks.

1.  **Build the binary:**
    ```bash
    go build -o netcfg-backup .
    ```

2.  **Managing the Device Inventory:**
    ```bash
    # List all configured devices
    ./netcfg-backup list

    # Add a new device interactively
    ./netcfg-backup add

    # Edit a device fields
    ./netcfg-backup edit

    # Remove a device by its host
    ./netcfg-backup remove <hostname_or_ip>
    ```

3.  **Running the Inventory Backup Process:**
    ```bash
    ./netcfg-backup run --backup-path /path/to/your/backups
    ```
    
4.  **Ad-hoc Command Execution:**
    Use the `exec` command to run commands on a single device without saving it to the inventory. All parameters are passed via flags.

    *Example (SSH with key):*
    ```bash
    ./netcfg-backup exec \
      --host 127.0.0.1 \
      --username your_user \
      --key-path ~/.ssh/your_key \
      --command "show version"
    ```
    *Example (Telnet with password from .env):*
    ```bash
    ./netcfg-backup exec \
      --host 10.0.0.1 \
      --username admin \
      --protocol telnet \
      --password-env "TELNET_PASSWORD" \
      --prompt "#" \
      --command "show running-config"
    ```

## Roadmap

-   [ ] Add a `web UI` for managing devices and viewing backup history.
-   [ ] Add support for alerting on failed backups via Prometheus Alertmanager.
-   [ ] Add support for pushing backups to S3-compatible object storage.

## License

This project is licensed under the MIT License. See the `LICENSE` file for details.