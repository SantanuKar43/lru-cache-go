# Leru - LRU Cache TCP Server

A high-performance TCP server implementing a Least Recently Used (LRU) cache with Write-Ahead Logging (WAL) for durability. Supports up to 1,000,000 cache entries with configurable TTL and flush strategies.

## Features

- **LRU Cache**: Stores up to N elements (max: 1,000,000) using Least Recently Used eviction policy
- **TTL Support**: Set expiration times for cache entries
- **Write-Ahead Logging (WAL)**: Ensures data durability with configurable flush strategies
- **TCP Server**: High-performance TCP interface for cache operations
- **Flexible Configuration**: Environment variable-based configuration
- **Docker Support**: Containerized deployment
- **Kubernetes Ready**: Includes deployment manifests

## Supported Commands

### PUT - Write to Cache
```
PUT {KEY: str} {VALUE: str}
PUT {KEY: str} {VALUE: str} {TTL_SECONDS: int}
```
Stores a key-value pair in the cache. Optionally specify a TTL (Time To Live) in seconds.

### GET - Read from Cache
```
GET {KEY: str}
```
Retrieves the value associated with a key from the cache.

### DEL - Delete from Cache
```
DEL {KEY: str}
```
Removes a key-value pair from the cache.

## Configuration

Configure the server using environment variables:

| Variable | Description | Example |
|----------|-------------|---------|
| `LERU_PORT_NUMBER` | Port number to listen for connections | `9090` |
| `LERU_CAPACITY` | Maximum number of cache elements | `10000` |
| `LERU_WAL_SIZE` | Maximum WAL file size in bytes | `1000000` |
| `LERU_FLUSH_STRATEGY` | Flush strategy: `SYNC` or `ASYNC` | `ASYNC` |

### Flush Strategies

- **SYNC**: High reliability mode. Each command is written to disk and fsync is called before responding to the client. Use for critical data.
- **ASYNC**: High performance mode. Commands are written to disk asynchronously. Better throughput but potential data loss on sudden failure.

## Build and Run

### Local Build
```bash
make build
./bin/leru
```

### Docker
```bash
make docker-run
```

### Kubernetes
```bash
# Deploy
make k8s-deploy-local

# Forward port to localhost
make k8s-port-forward

# Delete deployment
make k8s-delete-local
```

Refer to the [Makefile](./Makefile) for all available commands.

## Write-Ahead Logging (WAL)

The WAL is a command log file that records all operations for durability and recovery.

### WAL Behavior

- Each command is appended to `wal.log` with the configured flush strategy
- **SYNC mode**: fsync is called after each write, ensuring durability before responding to client
- **ASYNC mode**: Writes are buffered and flushed asynchronously for better performance

### Log Compaction

When the WAL file reaches the configured size limit, compaction occurs:

1. A new log file `wal.log.new` is created with all PUT commands for current cache entries
2. The order of entries is preserved to maintain LRU semantics
3. Old log file is renamed to `wal.log.old`
4. New log file becomes active (`wal.log`)
5. Old log file is deleted

### Recovery Process

On server restart, the WAL recovery proceeds as follows:

1. If both `wal.log` and `wal.log.new` exist:
   - `wal.log.new` is removed (incomplete compaction)
2. Else if only `wal.log.new` exists:
   - `wal.log.new` is renamed to `wal.log` (compaction completed)
3. If `wal.log.old` exists:
   - It is removed
4. The cache is rebuilt by replaying commands from `wal.log`
5. If `wal.log` doesn't exist, an empty one is created
6. The log is compacted if needed

## Architecture

The application is organized into the following modules:

- **cache**: Core LRU cache implementation with WAL support
- **config**: Configuration management (environment variables)
- **connection**: TCP connection handling and command processing
- **main**: Server initialization and connection acceptance

## Requirements

- Go 1.16 or higher
- Docker (for containerized deployment)
- kubectl (for Kubernetes deployment)