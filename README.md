# 📸 Image Compression Service (Go)

[![Go Report Card](https://goreportcard.com/badge/github.com/andreychano/compressor-golang)](https://goreportcard.com/report/github.com/andreychano/compressor-golang)
[![GitHub Release](https://img.shields.io/github/v/release/andreychano/compressor-golang?style=flat-square)](https://github.com/andreychano/compressor-golang/releases)
[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg?style=flat-square)](LICENSE)
[![Docker Pulls](https://img.shields.io/docker/pulls/andreychano/compressor-golang?style=flat-square)](https://hub.docker.com/r/andreychano/compressor-golang)

## 📖 Overview

A high‑performance microservice for **image compression and format conversion**, written in Go.  
The project follows **Hexagonal (Ports & Adapters) architecture** and uses **libvips** via the `bimg` wrapper for fast, low‑memory image processing.

## 🗂 Table of Contents

- [Features](#-features)
- [Architecture](#-architecture)
- [Quick Start](#-quick-start)
- [Configuration](#-configuration)
- [HTTP API](#-http-api)
- [Using as a Go Library](#-using-as-a-go-library)
- [Development & Testing](#-development--testing)
- [Contributing](#-contributing)
- [License](#-license)

## 🚀 Features

| ✅ | Description |
|---|---|
| **Dual operation modes** | *Storage Mode* – compress & persist to disk.<br>*Streaming Mode* – compress in‑memory and return the result instantly. |
| **Format conversion** | Supports JPEG, PNG, and WEBP. |
| **Security hardening** | Path‑traversal protection for file downloads. |
| **Clean architecture** | Business logic lives in `internal/core`, completely isolated from frameworks and third‑party libraries. |
| **Structured logging** | Correlation IDs, request sizes, client IPs, and error details are logged in JSON. |
| **Config‑driven** | All runtime behavior is controlled via `config.yaml`. |
| **Docker ready** | Multi‑stage Dockerfile for easy containerisation. |

## 🏗 Architecture

The repository follows the **Standard Go Project Layout** with a clear separation between adapters, core domain logic, and configuration.

```
/
├── INSTALL.md            # System‑level dependencies (libvips, build tools)
├── Makefile              # Convenient tasks: deps, build, run, test
├── README.md             # 📚 This file
├── cmd/
│   └── api/
│       └── main.go       # HTTP server entry point
├── compressor/
│   └── api.go            # Public façade for library usage
├── internal/
│   ├── adapter/
│   │   ├── inbound/
│   │   │   └── http/     # HTTP handlers + middleware
│   │   └── outbound/
│   │       ├── processor/
│   │       │   └── bimg/ # libvips implementation
│   │       └── repository/
│   │           └── local/ # Filesystem storage & path validation
│   ├── config/
│   │   ├── config.go     # Config structs
│   │   └── loader.go     # YAML loader
│   ├── core/
│   │   ├── domain/
│   │   │   └── file.go   # Domain models (File, Options, etc.)
│   │   ├── port/
│   │   │   ├── processor.go   # Processor port
│   │   │   └── repository.go  # Repository port
│   │   └── service/
│   │       └── compression.go # Business use‑cases
│   └── logger/
│       └── logger.go    # Zap‑based structured logger
├── config.yaml           # Default configuration (dev/prod overrides)
├── go.mod / go.sum
└── bin/
    └── api               # Compiled binary
```

## ⚡ Quick Start

### Prerequisites

- **Linux** (Ubuntu/Debian) – the project relies on native `libvips`.  
- `libvips` (≥ 8.9) – install via package manager.  
- **Go 1.25.5+** – the module uses recent language features.

```bash
# System dependencies (Ubuntu/Debian)
sudo apt-get update && sudo apt-get install -y libvips-dev build-essential

# Clone the repo
git clone https://github.com/pesiki-sobachki/compressor-golang.git
cd compressor-golang
```

### Build & Run (Make)

```bash
# Install Go dependencies & compile the binary
make deps      # go mod tidy + download libvips bindings
make build     # produces ./bin/api with config.local.yaml

# Run the server (default config.local.yaml → port 8080)
make run-local             #Run app in local mode with .env
```


## ⚙️ Configuration

Configuration lives in `internal/config/config.local.yaml`. Key sections:

```yaml
http:
  address: ":8080"
  max_upload_size_mb: 20
  read_timeout: "10s"
  write_timeout: "15s"
  idle_timeout: "60s"

storage:
  path: "./storage"
  compressed_subdir: "compressed"
  tmp_subdir: "tmp"

logger:
  level: "info" #level of logger
  service: "compressor-local" #servise name
  console: true #console output
  udp_address: "127.0.0.1:1515" #UDP address for logging
  enable_caller: false # Enable caller info in logs

image:
  default_format: "jpeg"
  default_quality: 50
  max_width: 3840
  max_height: 2160
  allow_formats: ["jpeg", "png", "webp"]
```

- **HTTP** – port, upload limit, and timeout settings.  
- **Storage** – root folder and sub‑folders for temporary and compressed files.  
- **Logger** – JSON output to console (or optional UDP collector).  
- **Image** – defaults for format, quality, and size constraints.

## 🌐 HTTP API

The service is reachable at `http://localhost:8080`.

### 1. Upload & Store (`POST /upload`)

Compresses an image and saves it to disk.

| Form field | Required | Description |
|------------|----------|-------------|
| `file` | ✅ | Binary image file (multipart). |
| `format` | ❌ | `jpeg` | `png` | `webp` (default from config). |
| `quality` | ❌ | 1‑100 (default from config). |

**cURL example**

```bash
curl -X POST http://localhost:8080/upload \
  -F "file=@/path/to/photo.jpg" \
  -F "format=webp" \
  -F "quality=80"
```

**Response**

```json
{
  "status": "success",
  "compressed_path": "storage/compressed/<uuid>.jpeg",
  "message": "File saved successfully"
}
```

### 2. Stream Compression (`POST /process`)

Compresses in‑memory and streams the result back.

```bash
curl -X POST http://localhost:8080/process \
  -F "file=@/path/to/photo.jpg" \
  -F "format=png" \
  -F "quality=80" \
  --output result.png
```

The response contains the binary image with appropriate `Content‑Type`, `Content‑Length` and `Content‑Disposition` headers.

### 3. Download (`GET /file?path=<relative_path>`)

Retrieves a previously stored file.

```bash
curl -v "http://localhost:8080/file?path=storage/compressed/<uuid>.jpeg" \
  --output downloaded.jpeg
```

- **400 Bad Request** – invalid or unsafe path.  
- **404 Not Found** – file missing or access denied.

## 📦 Using the Service as a Go Library

The same core can be imported directly:

```go
package main

import (
    "os"

    "github.com/pesiki-sobachki/compressor-golang/compressor"
)

func main() {
    // Initialise with default storage location
    comp := compressor.NewDefault("./storage")

    // Open source image
    src, err := os.Open("input.jpg")
    if err != nil {
        panic(err)
    }
    defer src.Close()

    // Compression options
    opts := compressor.Options{
        Format:   "webp",
        Quality:  80,
        MaxWidth: 0, // no width limit
        MaxHeight: 0,
    }

    // Perform compression
    data, meta, err := comp.Compress(src, opts)
    if err != nil {
        panic(err)
    }

    // Save result
    if err := os.WriteFile("output.webp", data, 0o644); err != nil {
        panic(err)
    }

    _ = meta // meta.MimeType, meta.Size, etc.
}
```

## 🧪 Development & Testing

```bash
# Run unit tests
make test

# Lint & format
make lint
```

## 🙋‍♀️ Contributing

1. Fork the repository.  
2. Create a feature branch (`git checkout -b feat/awesome`).  
3. Write tests for your changes.  
4. Ensure `make lint && make test` passes.  
5. Open a Pull Request describing the change.

Please adhere to the **Code of Conduct** and **conventional commit** style.

## 📜 License

Distributed under the **MIT License**. See `LICENSE` for details.

--- 

