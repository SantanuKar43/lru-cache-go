package cache

import (
	"bufio"
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"
)

const WAL_FILE = "leru-data/data/wal.log"
const WAL_FILE_NEW = "leru-data/data/wal.log.new"
const WAL_FILE_OLD = "leru-data/data/wal.log.old"
const WAL_PUT_LINE = "PUT %s %s %d %d\n"
const WAL_GET_LINE = "GET %s\n"
const WAL_DEL_LINE = "DEL %s\n"
var walMutex sync.Mutex

func AppendToWal(command string, c *Cache) error {
	switch (c.flushStrategy) {
	case SYNC:
		return appendSync(command, c.walSizeLimit, c)
	case ASYNC:
		go appendSync(command, c.walSizeLimit, c)
	}
	return nil
}

func appendSync(command string, walSizeLimit int64, cache *Cache) error {
	walMutex.Lock()
	defer walMutex.Unlock()
	file, err := os.OpenFile(WAL_FILE, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer file.Close()
	if _, err := file.WriteString(command); err != nil {
		return err
	}
	file.Sync()
	go compactWAL(walSizeLimit, cache)
	return nil
}

func RebuildCacheFromWAL(cache *Cache) {
	walMutex.Lock()
	defer walMutex.Unlock()
	if fileExists(WAL_FILE) {
		if fileExists(WAL_FILE_NEW) {
			os.Remove(WAL_FILE_NEW)
		}
	} else {
		if fileExists(WAL_FILE_NEW) {
			os.Rename(WAL_FILE_NEW, WAL_FILE)
		}
	}
	if fileExists(WAL_FILE_OLD) {
		os.Remove(WAL_FILE_OLD)
	}

	if fileExists(WAL_FILE) {
		file, err := os.Open(WAL_FILE)	
		if err != nil {
			log.Fatalf("Error opening file: %s", err)
		}
		scanner := bufio.NewScanner(file)

		for scanner.Scan() {
			// Get the current line as a string
			// TODO need to make sure line will fit in memory. 
			// Ideally it should as the line did fit in memory when the command was written.
			command := scanner.Text() 
			execute(command, cache)
		}
		if err := scanner.Err(); err != nil {
			log.Fatalf("Error scanning file: %s", err)
		}
	} else {
		createWALFile()
	}
	go compactWAL(cache.walSizeLimit, cache)
}

func createWALFile() {
	dir := filepath.Dir(WAL_FILE)
	err := os.MkdirAll(dir, 0755)
	if err != nil {
		log.Fatalf("Error creating WAL file: %s", err)
	}
	err = os.WriteFile(WAL_FILE, []byte(""), 0644)
	if err != nil {
		log.Fatalf("Error creating WAL file: %s", err)
	}
	fmt.Printf("Created file: %s\n", WAL_FILE)
}

func compactWAL(walSizeLimit int64, cache *Cache) {
	walMutex.Lock()
	defer walMutex.Unlock()
	fileInfo, err := os.Stat(WAL_FILE)
	if err != nil {
		fmt.Printf("Error getting file info: %v\n", err)
		return
	}
	if fileInfo.Size() > walSizeLimit {
		writeCacheToNewWAL(cache)
		os.Rename(WAL_FILE, WAL_FILE_OLD)
		os.Rename(WAL_FILE_NEW, WAL_FILE)
		os.Remove(WAL_FILE_OLD)
	}
}

func writeCacheToNewWAL(cache *Cache) {
	file, err := os.Create(WAL_FILE_NEW)
	if err != nil {
		fmt.Println("Error creating file:", err)
		return
	}
	defer file.Close()
	if cache.flushStrategy == ASYNC {
		cache.mutex.Lock()
		defer cache.mutex.Unlock()
	}
	elem := cache.lruCache.linkedList.Back()
	for elem != nil {
		node := elem.Value.(*Node)
		prev := elem.Prev()
		if !node.Expired() {
			fmt.Fprintf(file, WAL_PUT_LINE, node.Key, node.Val, node.CreatedAt, node.Ttl)
		}
		elem = prev
	}
	err = file.Sync()
	if err != nil {
		log.Fatal(err)
	}
}

func execute(input string, cache *Cache) {
	log := strings.Fields(input)
	switch (log[0]) {
	case "PUT":
		createdAt, _ := strconv.ParseInt(log[3], 10, 64) //ignoring error
		ttl, _ := strconv.ParseInt(log[4], 10, 64) //ignoring error
		timeElapsed := time.Now().Unix() - createdAt
		if ttl != 0 {
			if timeElapsed >= ttl {
				return
			} else {
				ttl -= timeElapsed
			}
		}
		cache.lruCache.Put(log[1], log[2], ttl)
	case "GET":
		cache.lruCache.Get(log[1])
	case "DEL":
		cache.lruCache.Delete(log[1])
	}

}

func fileExists(filename string) bool {
	info, err := os.Stat(filename)
	if errors.Is(err, os.ErrNotExist) {
		return false
	}
	// Check for other errors or if it's a directory (optional)
	return err == nil && !info.IsDir()
}