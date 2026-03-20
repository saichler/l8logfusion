# L8LogFusion

**Part of the Layer 8 Ecosystem**

A distributed log aggregation and fusion system for collecting, centralizing, and browsing logs from multiple sources in containerized environments.

[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](http://www.apache.org/licenses/LICENSE-2.0)
[![Go Version](https://img.shields.io/badge/Go-1.25.4-00ADD8.svg)](https://go.dev/)

## Overview

L8LogFusion is a Go-based microservice system that provides real-time log collection, centralized storage, and web-based browsing capabilities. It aggregates logs from distributed nodes over a Layer 8 virtual network overlay and consolidates them into a unified, queryable store.

## Features

- **Real-time Log Collection**: Monitors and tails log files with intelligent data-size-based buffering (512KB cap, 1s poll, 2s cooldown, 30s max age)
- **Active File Detection**: Uses Linux `/proc` filesystem analysis to discover which log files are currently being written to by other processes
- **Distributed Architecture**: Service-oriented design with virtual network overlay (L8Bus) for inter-component communication
- **File Rotation Handling**: Detects log file truncation/rotation via size comparison and seamlessly reopens
- **Web UI**: Browse and query collected logs through a web interface with file tree navigation
- **Pagination Support**: Efficient 5KB-per-page retrieval of large log files
- **Containerized Deployment**: Docker-ready components for cloud-native environments
- **Protocol Buffers**: Type-safe, efficient data serialization
- **Crash Recovery**: Resume tailing from a specific byte offset after restart
- **Configurable Storage**: Log database location is configurable (defaults to `/data/logdb`)

## Architecture

L8LogFusion consists of three main components:

```
Collector Agent(s)              Log Server                 Web UI
┌──────────────────┐     ┌──────────────────────┐    ┌──────────────┐
│ Monitor log dirs │     │ Receive log batches   │    │ File tree    │
│ Tail .log/.err   │────>│ Persist to disk       │<───│ navigation   │
│ Buffer & batch   │L8Bus│ /data/logdb/{IP}/{fn} │REST│ Paginated    │
│ Detect rotations │     │ File handle caching   │    │ log viewing  │
│ /proc scanning   │     │ Paginated queries     │    │ Health check │
└──────────────────┘     └──────────────────────┘    └──────────────┘
```

### 1. Log Collector Agent

- Scans directories recursively for `.log` and `.err` files
- Tails files in real-time with rotation detection (size-based truncation check)
- Buffers lines in a 30,000-entry queue with 512KB byte cap
- Flushes on cooldown (2s after last line), max age (30s), or buffer cap
- Periodically rescans for new files every 30 seconds
- Uses `/proc/[pid]/fd/` and `/proc/[pid]/fdinfo/` to detect actively-written files
- Supports wildcard (`*`) or specific filename monitoring

### 2. Log Server

- Central service receiving log batches from collectors via L8Bus unicast
- Persists logs organized by source IP and filename under a configurable root directory
- Thread-safe file handle cache for optimized write I/O
- Provides REST GET endpoint for querying:
  - Directory listing (file tree structure)
  - Paginated file content (5KB per page)
- Supports multi-instance merge for distributed queries

### 3. Web UI

- Web-based interface for browsing collected logs
- File tree navigation across all collected sources
- Health monitoring via Probler integration
- Runs on port 26000 (HTTPS)

## Data Flow

```
1. Collector opens log file, seeks to position (start/end/offset)
2. Polls file every 1s, reads new lines into queue buffer
3. Flushes buffer when:
   - No new lines for 2s (cooldown)
   - Buffer age exceeds 30s
   - Buffer exceeds 512KB
   - File rotation detected (truncation)
4. Sends L8LogF{sourceIp, filename, logs[]} via L8Bus unicast to server
5. Server writes log bytes to /data/logdb/{sourceIp}/{filename}
6. Web UI queries server via REST GET with L8Query for tree or paginated content
```

## Installation

### Prerequisites

- Go 1.25.4 or higher
- Docker (for containerized deployment)
- Linux (for `/proc`-based active file detection)

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
cd ../../ui && ./build.sh
```

## Usage

### Running the Log Collector

```bash
# Set environment variables
export NODE_IP=192.168.1.100  # Source IP identifier (falls back to auto-detect)
export LOGPATH=/var/log        # Directory to monitor
export LOGFILE="*"             # "*" for all .log/.err files, or a specific filename

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
# Start the web UI (HTTPS on port 26000)
go run go/agent/ui/main.go
```

## Configuration

### Environment Variables

**Log Collector:**
- `NODE_IP`: IP address of the current node (falls back to machine IP auto-detection)
- `LOGPATH`: Directory path to monitor for logs
- `LOGFILE`: Specific filename or `*` for all `.log`/`.err` files

**Log Server:**
- Command-line argument: Virtual network port (e.g., `12443`)
- Storage location: Configurable via service activation (defaults to `/data/logdb`)

**Web UI:**
- HTTPS port: 26000
- TLS certificates: `/data/probler`

### Buffering Parameters

| Parameter | Value | Purpose |
|-----------|-------|---------|
| Poll interval | 1s | How often to check for new file data |
| Cooldown duration | 2s | Flush after this idle period |
| Max buffer age | 30s | Force flush regardless of activity |
| Max buffer bytes | 512KB | Force flush when buffer reaches this size |
| Queue capacity | 30,000 lines | Maximum buffered lines before blocking |

### Storage Layout

Logs are persisted with the following directory structure:
```
/data/logdb/
├── 192.168.1.100/
│   ├── application.log
│   ├── error.err
│   └── ...
├── 192.168.1.101/
│   └── ...
```

## Protocol Buffer Types

The system uses Protocol Buffers for data serialization:

```protobuf
// Primary log batch message (collector -> server)
message L8LogF {
    string sourceIp = 1;        // Source machine IP
    string filename = 2;        // Log file name
    repeated string logs = 3;   // Log lines
}

// Log source configuration
message L8LogConfig {
    string path = 1;            // Directory to monitor
    string name = 2;            // Filename or "*" for all
}

// File tree node (server -> web UI)
message L8File {
    bool isDirectory = 1;
    string name = 2;
    string path = 3;
    repeated L8File files = 4;  // Child entries
    L8FileData data = 5;        // Paginated content
}

// Pagination metadata
message L8FileData {
    int32 limit = 1;            // Bytes per page (5120)
    int32 page = 2;             // Current page (0-indexed)
    int32 size = 3;             // Total file size
    string content = 4;         // Page content
}
```

## Development

### Project Structure

```
l8logfusion/
├── proto/
│   ├── logs.proto              # Protocol Buffer definitions
│   └── make-bindings.sh        # Protobuf compilation script
├── go/
│   ├── agent/
│   │   ├── common/             # Service constants (name, area)
│   │   │   ├── common.go       # LogServiceName, LogServiceArea
│   │   │   └── files.go        # FileOf() directory tree builder
│   │   ├── logs/               # Log collector agent
│   │   │   ├── LogCollector.go # Orchestration & SendLogs
│   │   │   ├── Tail.go         # Core tailing with buffering
│   │   │   ├── FileTracker.go  # Duplicate tail prevention
│   │   │   ├── ProcScanner.go  # /proc-based active file detection
│   │   │   └── docker/main.go  # Docker entry point
│   │   ├── logserver/          # Central log server
│   │   │   ├── LogService.go   # Service handler (Post/Get/Merge)
│   │   │   ├── LogCenter.go    # Helpers (fetchFile, LoadData)
│   │   │   └── docker/main.go  # Docker entry point
│   │   └── ui/                 # Web UI
│   │       ├── main.go         # Entry point
│   │       ├── websvr/         # Web server setup
│   │       └── web/            # HTML, CSS, JS assets
│   ├── types/l8logf/           # Generated protobuf Go code
│   └── tests/                  # Integration tests
├── plans/                      # Development plans
└── LICENSE                     # Apache 2.0
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

| Image | Description |
|-------|-------------|
| `saichler/probler-logagent:latest` | Log collector agent |
| `saichler/probler-logserver:latest` | Central log server |
| `saichler/probler-logsui:latest` | Web UI |

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

| Dependency | Purpose |
|-----------|---------|
| [l8bus](https://github.com/saichler/l8bus) | Virtual network overlay for distributed communication |
| [l8types](https://github.com/saichler/l8types) | Shared interface definitions (IVNic, IServiceHandler) |
| [l8utils](https://github.com/saichler/l8utils) | Utilities (logger, IP detection, queue) |
| [l8web](https://github.com/saichler/l8web) | Web server framework (REST API) |
| [l8srlz](https://github.com/saichler/l8srlz) | Serialization support |
| [l8reflect](https://github.com/saichler/l8reflect) | Reflection-based introspection |
| [probler](https://github.com/saichler/probler) | Health monitoring and profiling |
| [Protocol Buffers](https://developers.google.com/protocol-buffers) | Data serialization |

## Layer 8 Ecosystem

L8LogFusion is part of the Layer 8 Ecosystem, a suite of tools for building distributed systems:

- [l8bus](https://github.com/saichler/l8bus) - Virtual network overlay
- [l8types](https://github.com/saichler/l8types) - Shared type definitions
- [l8utils](https://github.com/saichler/l8utils) - Utility functions
- [l8reflect](https://github.com/saichler/l8reflect) - Reflection-based introspection
- [l8srlz](https://github.com/saichler/l8srlz) - Serialization support
- [l8orm](https://github.com/saichler/l8orm) - Object-relational mapping
- [l8web](https://github.com/saichler/l8web) - Web server framework
- [probler](https://github.com/saichler/probler) - Monitoring and profiling

## Contributing

Contributions are welcome! Please feel free to submit pull requests or open issues for bugs and feature requests.

## License

Copyright 2025 Sharon Aicler (saichler@gmail.com)

Licensed under the Apache License, Version 2.0. You may obtain a copy of the License at:

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the specific language governing permissions and limitations under the License.

## Author

**Sharon Aicler** (saichler@gmail.com)

## Support

For issues, questions, or contributions, please visit the [GitHub repository](https://github.com/saichler/l8logfusion).
