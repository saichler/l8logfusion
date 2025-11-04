# L8LogFusion

A distributed log aggregation and fusion system for collecting, centralizing, and browsing logs from multiple sources in containerized environments.

## Overview

L8LogFusion is a Go-based microservice system that provides real-time log collection, centralized storage, and web-based browsing capabilities. It's designed to aggregate logs, traps, and events from distributed nodes and consolidate them into a unified, queryable stream.

## Features

- **Real-time Log Collection**: Monitors and collects logs from multiple sources with intelligent buffering
- **Distributed Architecture**: Service-oriented design with virtual network overlay for inter-component communication
- **File Rotation Handling**: Automatically detects and handles log file rotations
- **Web UI**: Browse and query collected logs through a web interface
- **Pagination Support**: Efficient retrieval of large log files with 5KB per page
- **Containerized Deployment**: Docker-ready components for cloud-native environments
- **Protocol Buffers**: Type-safe, efficient data serialization
- **Resilient Design**: Crash recovery with resume capability from last read position

## Architecture

L8LogFusion consists of three main components:

### 1. Log Collector Agent
- Scans directories for `.log` and `.err` files
- Tails files in real-time with rotation detection
- Buffers and batches log entries for efficient transmission
- Supports wildcard collection patterns

### 2. Log Server
- Central service that receives logs from collectors
- Persists logs organized by source IP and filename
- Provides REST API for querying and pagination
- Manages file handle caching for optimized I/O

### 3. Web UI
- Web-based interface for browsing logs
- Health monitoring endpoints
- File tree navigation
- Runs on port 26000 by default

## Installation

### Prerequisites
- Go 1.24.9 or higher
- Docker (for containerized deployment)
- Protocol Buffers compiler (for regenerating proto files)

### Building from Source

```bash
# Clone the repository
git clone https://github.com/saichler/l8logfusion
cd l8logfusion

# Run tests
cd go
./test.sh

# Build Docker images
cd agent/logs/docker && ./build.sh
cd ../../logserver/docker && ./build.sh
cd ../../ui/docker && ./build.sh
```

## Usage

### Running the Log Collector

```bash
# Set environment variables
export NODE_IP=192.168.1.100  # Or auto-detect
export LOGPATH=/var/log        # Directory to monitor
export LOGFILE="*"             # "*" for all .log/.err files or specific filename

# Run the collector
go run go/agent/logs/docker/main.go
```

### Running the Log Server

```bash
# Start the log server on virtual network port 12443
go run go/agent/logserver/docker/main.go 12443
```

### Running the Web UI

```bash
# Start the web UI (defaults to port 26000)
go run go/agent/ui/main.go
```

## Configuration

### Environment Variables

**Log Collector:**
- `NODE_IP`: IP address of the current node
- `LOGPATH`: Directory path to monitor for logs
- `LOGFILE`: Specific filename or "*" for all log files

**Log Server:**
- Command-line argument: Virtual network port (e.g., 12443)

**Web UI:**
- Runs on port 26000 (hardcoded)
- Data path: `/data/probler`

### Storage

Logs are persisted to `/data/logdb` with the following structure:
```
/data/logdb/
├── {SOURCE_IP}/
│   ├── {FILENAME_1}
│   ├── {FILENAME_2}
│   └── ...
```

## API Reference

### Log Server REST Endpoints

- `POST /logs` - Submit new log entries
- `GET /logs` - Query stored logs with pagination
- `PUT /logs` - Update existing logs
- `PATCH /logs` - Merge log entries
- `DELETE /logs` - Remove log entries

### Data Models

The system uses Protocol Buffers for data serialization. Key message types:

```protobuf
message L8LogF {
    string sourceIp = 1;        // IP of source machine
    string filename = 2;        // Log file name
    repeated string logs = 3;   // Array of log lines
}

message L8LogConfig {
    string path = 1;           // Directory to monitor
    string name = 2;           // Filename or "*" for all
}
```

## Development

### Project Structure

```
l8logfusion/
├── proto/                  # Protocol Buffer definitions
├── go/
│   ├── agent/
│   │   ├── logs/          # Log collector implementation
│   │   ├── logserver/     # Central log server
│   │   └── ui/            # Web UI
│   ├── types/             # Generated protobuf types
│   ├── tests/             # Test suite
│   └── common/            # Shared utilities
└── LICENSE                # Apache 2.0 License
```

### Running Tests

```bash
cd go
./test.sh  # Runs tests with coverage and security checks
```

### Regenerating Protocol Buffers

```bash
cd proto
./make-bindings.sh
```

## Docker Deployment

The project provides three Docker images:

- `saichler/probler-logagent:latest` - Log collector agent
- `saichler/probler-logserver:latest` - Central log server
- `saichler/probler-logsui:latest` - Web UI

Example Docker Compose configuration:

```yaml
version: '3.8'
services:
  logserver:
    image: saichler/probler-logserver:latest
    command: ["12443"]
    volumes:
      - /data/logdb:/data/logdb

  logagent:
    image: saichler/probler-logagent:latest
    environment:
      - NODE_IP=auto
      - LOGPATH=/var/log
      - LOGFILE=*
    volumes:
      - /var/log:/var/log:ro

  logsui:
    image: saichler/probler-logsui:latest
    ports:
      - "26000:26000"
    volumes:
      - /data/probler:/data/probler
```

## Dependencies

- [l8bus](https://github.com/saichler/l8bus) - Virtual network overlay
- [l8types](https://github.com/saichler/l8types) - Shared type definitions
- [l8utils](https://github.com/saichler/l8utils) - Utility functions
- [l8reflect](https://github.com/saichler/l8reflect) - Reflection-based introspection
- [l8srlz](https://github.com/saichler/l8srlz) - Serialization support
- [l8web](https://github.com/saichler/l8web) - Web server framework
- [probler](https://github.com/saichler/probler) - Monitoring/profiling
- [Protocol Buffers](https://developers.google.com/protocol-buffers) - Data serialization

## Contributing

Contributions are welcome! Please feel free to submit pull requests or open issues for bugs and feature requests.

## License

This project is licensed under the Apache License 2.0 - see the [LICENSE](LICENSE) file for details.

## Author

Shai Saichler

## Support

For issues, questions, or contributions, please visit the [GitHub repository](https://github.com/saichler/l8logfusion).
