package cache

import (
	"container/list"
)

type Node struct {
	Key string
	Val string
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
	c.linkedList.MoveToFront(elem)
	return elem.Value.(Node).Val
}

func (c *LRUCache) Put(key string, val string) {
	elem, exists := c.hashMap[key]
	if exists {
		c.linkedList.MoveToFront(elem)
		return
	}
	elem = c.linkedList.PushFront(Node{key, val})
	c.hashMap[key] = elem
	if c.linkedList.Len() > c.capacity {
		node := c.linkedList.Remove(c.linkedList.Back()).(Node)
		delete(c.hashMap, node.Key)
	}
}

