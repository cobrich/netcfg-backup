# Netcfg-Backup

A reliable and extensible tool written in Go for automated backup of network device configurations, featuring a built-in monitoring stack.

[![Go Report Card](https://goreportcard.com/badge/github.com/cobrich/netcfg-backup)](https://goreportcard.com/report/github.com/cobrich/netcfg-backup)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

`netcfg-backup` allows you to connect to a large number of network devices concurrently via SSH or Telnet, execute a predefined set of commands, and save the output. It comes with a pre-configured observability stack using Prometheus and Grafana to monitor job performance.

## Key Features

-   **Built-in Monitoring Stack:** Comes with a ready-to-use `docker-compose` setup for Prometheus and Grafana, providing instant insights into job performance, duration, and success rates.
-   **User-Friendly CLI:** Manage your device inventory with intuitive commands (`add`, `list`, `remove`, `run`).
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
    -   Use the interactive CLI to add your devices: `./netcfg-backup add`. This will create the `devices/devices.json` file.

## Usage

There are two main ways to run the application: as a full stack with monitoring, or as a standalone CLI tool.

### Running the Full Stack with Monitoring (Recommended)

This is the easiest way to get started and see all the features in action.

1.  **Launch all services:**
    ```bash
    docker compose up --build
    ```
    This command will build the application, and start the `netcfg-backup`, `prometheus`, and `grafana` containers.

2.  **View the metrics:**
    -   **Prometheus:** Open `http://localhost:9091` to see the Prometheus UI. You can check the status of the `netcfg-backup` target under `Status -> Targets`.
    -   **Grafana:** Open `http://localhost:3000`.
        -   Login with `admin` / `admin`.
        -   Add Prometheus as a data source (URL: `http://prometheus:9090`).
        -   Create a dashboard to visualize the `netcfg_backup_*` metrics.

### Standalone CLI Usage

You can also build and run the tool directly for quick tasks.

1.  **Build the binary:**
    ```bash
    go build -o netcfg-backup .
    ```

2.  **Manage devices:**
    ```bash
    # List all configured devices
    ./netcfg-backup list

    # Add a new device interactively
    ./netcfg-backup add

    # Remove a device by its host
    ./netcfg-backup remove <hostname_or_ip>
    ```

3.  **Run the backup process:**
    ```bash
    ./netcfg-backup run --backup-path /path/to/your/backups
    ```

## Roadmap

-   [ ] Add a `web UI` for managing devices and viewing backup history.
-   [ ] Implement an `edit` command to modify existing devices.
-   [ ] Add support for alerting on failed backups via Prometheus Alertmanager.
-   [ ] Add support for pushing backups to S3-compatible object storage.

## License

This project is licensed under the MIT License. See the `LICENSE` file for details.