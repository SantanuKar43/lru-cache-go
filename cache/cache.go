package cache

import (
	"bufio"
	"container/list"
	"errors"
	"fmt"
	"log"
	"os"
	"strings"
	"sync"
	"path/filepath"
)

type FlushStrategy string

const (
	SYNC FlushStrategy = "SYNC"
	ASYNC FlushStrategy = "ASYNC"
)

type Cache struct {
	mutex sync.Mutex
	cache *LRUCache
	walSizeLimit int64
	flushStrategy FlushStrategy
}

var walMutex sync.Mutex

const WAL_FILE = "leru/data/wal.log"
const WAL_FILE_NEW = "leru/data/wal.log.new"
const WAL_FILE_OLD = "leru/data/wal.log.old"

func Init(capacity int, walSizeLimit int64, flushStrategy FlushStrategy) *Cache {
	c := &Cache {
		cache: &LRUCache{
			hashMap: make(map[string]*list.Element),
			linkedList: list.New(),
			capacity: capacity,
		},
		walSizeLimit: walSizeLimit,
		flushStrategy: flushStrategy,
	}
	rebuildCacheFromWAL(c)
	return c
}

func (c *Cache) Get(key string) (string, error) {
	defer c.mutex.Unlock()
	c.mutex.Lock()
	val := c.cache.Get(key)
	if c.flushStrategy == SYNC {
		err := appendToWal(fmt.Sprintf("GET %s\n", key), c.walSizeLimit, c.cache)
		return val, err
	} else {
		go appendToWal(fmt.Sprintf("GET %s\n", key), c.walSizeLimit, c.cache)
	}
	return val, nil
}

func (c *Cache) Put(key string, val string) error {
	defer c.mutex.Unlock()
	c.mutex.Lock()
	c.cache.Put(key, val)
	if c.flushStrategy == SYNC {
		err := appendToWal(fmt.Sprintf("PUT %s %s\n", key, val), c.walSizeLimit, c.cache)
		return err
	} else {
		go appendToWal(fmt.Sprintf("PUT %s %s\n", key, val), c.walSizeLimit, c.cache)
	}
	return nil
}

func appendToWal(command string, walSizeLimit int64, lruCache *LRUCache) error {
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
	go compactWAL(walSizeLimit, lruCache)
	return nil
}

func rebuildCacheFromWAL(cache *Cache) {
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
		// Create a new scanner for the file
		scanner := bufio.NewScanner(file)

		// Iterate over the scanner's lines
		for scanner.Scan() {
			// Get the current line as a string
			// TODO need to make sure line will fit in memory. 
			// Ideally it should as the line did fit in memory when the command was written.
			command := scanner.Text() 
			execute(command, cache)
		}

		// Check for errors during scanning
		if err := scanner.Err(); err != nil {
			log.Fatalf("Error scanning file: %s", err)
		}
	} else {
		createWALFile()
	}
	go compactWAL(cache.walSizeLimit, cache.cache)
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

func compactWAL(walSizeLimit int64, lruCache *LRUCache) {
	walMutex.Lock()
	defer walMutex.Unlock()
	fileInfo, err := os.Stat(WAL_FILE)
	if err != nil {
		fmt.Printf("Error getting file info: %v\n", err)
		return
	}
	if fileInfo.Size() > walSizeLimit {
		writeCacheToNewWAL(lruCache)
		os.Rename(WAL_FILE, WAL_FILE_OLD)
		os.Rename(WAL_FILE_NEW, WAL_FILE)
		os.Remove(WAL_FILE_OLD)
	}
}

func writeCacheToNewWAL(lruCache *LRUCache) {
	file, err := os.Create(WAL_FILE_NEW)
	if err != nil {
		fmt.Println("Error creating file:", err)
		return
	}
	defer file.Close()

	elem := lruCache.linkedList.Back()
	for elem != nil {
		fmt.Fprintf(file, "PUT %s %s\n", elem.Value.(Node).Key, elem.Value.(Node).Val)
		elem = elem.Prev()
	}
	err = file.Sync()
	if err != nil {
		log.Fatal(err)
	}
}

func execute(input string, cache *Cache) {
	command := strings.Fields(input)
	switch (command[0]) {
	case "PUT":
		cache.cache.Put(command[1], command[2])
	case "GET":
		cache.cache.Get(command[1])
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
