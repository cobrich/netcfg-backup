```markdown
# Netcfg-Backup

A reliable and extensible tool written in Go for automated backup of network device configurations.

[![Go Report Card](https://goreportcard.com/badge/github.com/cobrich/netcfg-backup)](https://goreportcard.com/report/github.com/cobrich/netcfg-backup)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

`netcfg-backup` allows you to connect to a large number of network devices concurrently via SSH or Telnet, execute a predefined set of commands (e.g., `show running-config`), and save the output to structured, timestamped files.

## Key Features

-   **Multi-protocol Support:** Connect to devices using SSH (with key-based authentication) or legacy Telnet.
-   **Concurrent Operations:** Efficiently handles large device lists using a worker pool to control the load on your network and the application host.
-   **Flexible Configuration:** Manage the device inventory, credentials, and commands through a simple JSON file.
-   **Secure Credential Management:** Passwords and other secrets are handled securely via environment variables, with support for `.env` files for easy local development.
-   **Organized Backups:** Command outputs are saved to separate, timestamped files for each device, creating a clear audit trail.
-   **DevOps Ready:** Features structured JSON logging and includes a multi-stage `Dockerfile` for quick and secure containerization.

## Getting Started

### Prerequisites

-   [Go](https://golang.org/doc/install) (version 1.24 or later)
-   [Docker](https://docs.docker.com/get-docker/) (for containerized deployment)
-   Git

### Installation

1.  Clone the repository:
    ```bash
    git clone https://github.com/cobrich/netcfg-backup.git
    cd netcfg-backup
    ```

2.  Build the application binary:
    ```bash
    go build -o netcfg-backup .
    ```

## Configuration

1.  **Device Inventory:** Copy the example `devices/devices.json.example` to `devices/devices.json` and customize it with your list of devices.
    -   For SSH key authentication, use the `"key_path"` field.
    -   For password authentication (Telnet or SSH), use the `"password_env"` field to specify the name of an environment variable.

2.  **Secrets:** Copy the example `.env.example` to `.env` and provide the values for the environment variables you defined in your `devices.json`.

    **Important:** The `.env` file should never be committed to version control. It is already included in `.gitignore`.

## Usage

### Local Execution

Run the application from the project root. Ensure your `.env` file is populated.
```bash
./netcfg-backup
```

You can specify a custom backup location using the `--backup-path` flag:
```bash
./netcfg-backup --backup-path /var/backups/network-configs
```

### Running with Docker

1.  Build the Docker image:
    ```bash
    docker build -t netcfg-backup:latest .
    ```

2.  Run the container, mounting the necessary files and directories from your host machine:
    ```bash
    docker run --rm \
      --env-file ./.env \
      -v "$(pwd)/devices:/app/devices:ro" \
      -v "$(pwd)/backups:/app/backups" \
      -v "$(pwd)/logs:/app/logs" \
      -v "$HOME/.ssh/your_ssh_key:/root/.ssh/your_ssh_key:ro" \
      -v "$HOME/.ssh/known_hosts:/root/.ssh/known_hosts:ro" \
      netcfg-backup:latest \
      --backup-path ./backups
    ```

## Roadmap

-   [ ] Add support for fetching configurations via SNMP and NETCONF.
-   [ ] Implement output parsers to store data in structured formats (JSON/CSV).
-   [ ] Integrate with Prometheus for application and job metrics.
-   [ ] Add support for pushing backups to S3-compatible object storage.

## License

This project is licensed under the MIT License. See the `LICENSE` file for details.
```

Теперь, после сохранения этого файла и отправки на GitHub, форматирование будет идеальным. Спасибо за вашу внимательность