## TCP based LRU cache server - Leru
Stores N (max: 1000000) least recently used elements.

Uses a WAL for durability with configurable flush strategy - always for each command, or every 1 second.

### Configuration:
Configuration is supplied using env variables
```
LERU_PORT: port number to listen for connections
LERU_CAPACITY: number of elements to store
LERU_WAL_SIZE: size limit of WAL in bytes
LERU_FLUSH_STRATEGY: SYNC / ASYNC
```
### Supported commands:
#### Write: Put in cache
`PUT {KEY: str} {VALUE: str}`

#### Read: Get from cache
`GET {KEY: str}`

## WAL usage
WAL is a simple log file containing commands received by the server.

When a command is received, it is appended to the log file with the configured flush strategy.

Flush strategy "SYNC" - This is used for high reliability. The log line is appended in sync. Response is returned after fsync call returns.

Flush strategy "ASYNC" - This is used for high performance. The log line is appended in async.

### Compaction
When log file size limit is reached, the cache is examined and the log file is replaced with PUT commands, written for each element in cache preserving the order in memory.

#### Steps:
1. WAL log file name: `wal.log`
2. When size limit reached, a `wal.log.new` file is created with new logs. 
3. `wal.log` is renamed to `wal.log.old`
4. `wal.log.new` is renamed to `wal.log`
5. `wal.log.old` is removed.

On restart,
1. If `wal.log` file is present:
   - If `wal.log.new` file is present:
     - `wal.log.new` is removed.
2. else if `wal.log.new` is present
   - `wal.log.new`is renamed to `wal.log`.
3. If `wal.log.old` file is present, it is removed.
4. `wal.log` file is referenced to rebuild the cache. If `wal.log` is not present, an empty `wal.log` is created.
5. `wal.log` is compacted if needed.
