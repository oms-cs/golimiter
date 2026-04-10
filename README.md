# golimiter

A distributed rate limiting service built with Go, Redis, and gRPC. It provides multiple rate limiting algorithms for microservices architectures.

## Features

- **Multiple Algorithms**: Token Bucket, Sliding Window Counter, Sliding Window Log
- **Distributed**: Redis-backed for horizontal scalability
- **gRPC API**: High-performance interface with protobuf
- **Configurable**: Support for multiple rate limits per service/path
- **Production Ready**: Atomic Redis operations with Lua scripts

## Configuration

The service uses YAML configuration to define rate limits per service and path:

```yaml
resources:
  - service: "product-service"
    algorithm: sliding_window_log
    paths:
      - path: "/products"
        method: "GET"
        rules:
          limits:
            - window_seconds: 30
              limit: 10
            - window_seconds: 60
              limit: 12
      - path: "/products/:id"
        method: "GET"
        rules:
          limits:
            - window_seconds: 10
              limit: 10
            - window_seconds: 60
              limit: 50
            - window_seconds: 86400
              limit: 1000
```

## Rate Limiting Algorithms

- **Token Bucket**: Allows bursts, refills at constant rate
- **Sliding Window Counter**: Two windows with weighted decay
- **Sliding Window Log**: Tracks individual timestamps (most accurate)

## Quick Start

```bash
make install  # Install dependencies
make generate # Generate protobuf code
make build    # Build binary
make run      # Start service
```

## Requirements

- Go 1.25+
- Redis 6.0+
- Protocol Buffers compiler (protoc)
