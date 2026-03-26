# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Overview

Backend for Wuxi University - Novosibirsk College education evaluation system. A Go-based microservice using the go-kratos framework that supports high-concurrency student evaluations of teaching (~1000 QPS target).

## Project Information

- **Language**: Go 1.22+
- **Framework**: Go-Kratos v2
- **Database**: MySQL/PostgreSQL (GORM)
- **Cache**: Redis
- **API**: gRPC + HTTP
- **Dependency Injection**: Google Wire
- **Code Generation**: Protobuf plugins

## Commands

### Initialization
```bash
make init          # Install required protoc plugins and tools
```

### Code Generation
```bash
make all           # Generate API proto, config proto, and run go generate
make api           # Generate Go code from API proto files (HTTP/gRPC/OpenAPI)
make config        # Generate Go code from internal config proto
make generate      # Run go generate and tidy modules
```

### Building
```bash
make build         # Build the binary to ./bin/
go build -o ./bin/ ./...  # Alternative build command
```

### Running
```bash
./bin/edu-evaluation-backed -conf configs/config.yaml  # Run the server
```

### Testing
```bash
go test ./...      # Run all tests
go test ./path/to/package  # Run single package tests
```

### Dependencies
```bash
go mod tidy        # Update module dependencies
```

## Architecture

### Clean Architecture / Layered Structure

```
cmd/                    - Application entry point and wire configuration
├── main.go            - Main entry: loads config, starts app
└── wire.go            - Wire DI configuration

internal/
├── conf/              - Configuration (protobuf-defined)
├── server/            - Transport layer setup (gRPC and HTTP servers)
├── service/           - Business API endpoints (use case orchestration)
├── biz/               - Business logic domain layer
├── data/              - Data access layer (repository implementations)
│   └── common/        - Common data utilities
│       ├── gorm/      - GORM database initialization
│       └── redis/     - Redis client initialization
└── common/            - Common utilities (constants, utils)
    └── utils/auth/    - Authentication utilities

configs/               - Configuration files
third_party/           - Third-party proto dependencies (google, validate, etc.)
openapi.yaml           - Generated OpenAPI specification
```

### Dependency Injection

Uses Google Wire for compile-time DI. The dependency graph is assembled in `cmd/wire.go` and generated into `cmd/wire_gen.go`. Run `make generate` after changing provider sets.

### Configuration

Configuration is defined via protobuf (`internal/conf/conf.proto`) and loaded from YAML files at startup (`configs/config.yaml`). The structure includes:
- Server HTTP/gRPC listen addresses and timeouts
- Data source configuration (database driver, DSN, Redis settings)

### Data Layer

- Supports both Postgres (currently implemented) and MySQL (configured in YAML)
- Auto-migration is enabled in `data/data.go`
- Database connection handled by GORM
- Redis connection for caching

### Transport

The server exposes both:
- HTTP on port 8000
- gRPC on port 9000

## Technology Stack

Backend: Golang + go-kratos + GORM + Redis + MySQL/PostgreSQL
Frontend: Uniapp + TypeScript + Vue (separate repository)

## Business Domain

Student evaluation system for courses/instructors with two main portals:
- **Student app**: Login, view courses/instructors, submit evaluations
- **Admin portal**: Login, manage classes/students/courses/instructors, view/export evaluation results
