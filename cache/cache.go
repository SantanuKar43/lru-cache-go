package cache

import (
	"container/list"
	"fmt"
	"sync"
)

type FlushStrategy string

const (
	SYNC FlushStrategy = "SYNC"
	ASYNC FlushStrategy = "ASYNC"
)

type Cache struct {
	mutex sync.Mutex
	lruCache *LRUCache
	walSizeLimit int64
	flushStrategy FlushStrategy
}

func Init(capacity int, walSizeLimit int64, flushStrategy FlushStrategy) *Cache {
	c := &Cache {
		lruCache: &LRUCache{
			hashMap: make(map[string]*list.Element),
			linkedList: list.New(),
			capacity: capacity,
		},
		walSizeLimit: walSizeLimit,
		flushStrategy: flushStrategy,
	}
	RebuildCacheFromWAL(c)
	return c
}

func (c *Cache) Get(key string) (string, error) {
	defer c.mutex.Unlock()
	c.mutex.Lock()
	val := c.lruCache.Get(key)
	AppendToWal(fmt.Sprintf(WAL_GET_LINE, key), c)
	return val, nil
}

func (c *Cache) Put(key string, val string, ttl int64) error {
	defer c.mutex.Unlock()
	c.mutex.Lock()
	node := c.lruCache.Put(key, val, ttl)
	AppendToWal(fmt.Sprintf(WAL_PUT_LINE, key, val, node.CreatedAt, ttl), c)
	return nil
}

func (c *Cache) Delete(key string) error {
	defer c.mutex.Unlock()
	c.mutex.Lock()
	c.lruCache.Delete(key)
	AppendToWal(fmt.Sprintf(WAL_DEL_LINE, key), c)
	return nil
}
