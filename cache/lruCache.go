package cache

import (
	"container/list"
	"time"
)

type Node struct {
	Key string
	Val string
	CreatedAt int64
	Ttl int64
}

func (node *Node) Expired() bool {
	return node.Ttl != 0 && ((node.CreatedAt + node.Ttl) > time.Now().Unix())
}

type LRUCache struct {
	hashMap map[string]*list.Element
	linkedList *list.List
	capacity int
}

func (c *LRUCache) Get(key string) string {
	elem, exists := c.hashMap[key]
	if !exists {
		return ""
	}
	node := elem.Value.(Node)
	if node.Expired() {
		c.Delete(key)
		return ""
	}
	c.linkedList.MoveToFront(elem)
	return elem.Value.(Node).Val
}

func (c *LRUCache) Put(key string, val string, ttl int64) Node {
	elem, exists := c.hashMap[key]
	if exists {
		c.linkedList.MoveToFront(elem)
		return elem.Value.(Node)
	}
	elem = c.linkedList.PushFront(Node{key, val, time.Now().Unix(), ttl})
	c.hashMap[key] = elem
	if c.linkedList.Len() > c.capacity {
		c.Delete(c.linkedList.Back().Value.(Node).Key)
	}
	return elem.Value.(Node)
}

func (c *LRUCache) Delete(key string) {
	elem, exists := c.hashMap[key]
	if !exists {
		return
	}
	node := c.linkedList.Remove(elem).(Node)
	delete(c.hashMap, node.Key)
}

