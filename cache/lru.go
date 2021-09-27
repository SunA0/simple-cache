package cache

import "container/list"

type lruCache struct {
	maxBytes  int64
	nBytes    int64
	ll        *list.List
	cache     map[string]*list.Element
	OnEvicted func(key string, val Value)
}

type entry struct {
	key string
	val Value
}

type Value interface {
	Len() int
}

func NewLru(maxBytes int64, onEvicted func(string, Value)) *lruCache {
	return &lruCache{
		maxBytes:  maxBytes,
		ll:        list.New(),
		cache:     make(map[string]*list.Element),
		OnEvicted: onEvicted,
	}
}

func (c *lruCache) Add(key string, val Value) {
	// add
	if ele, ok := c.cache[key]; ok {
		c.ll.MoveToFront(ele)
		kv := ele.Value.(entry)
		c.nBytes += int64(val.Len()) - int64(kv.val.Len())
		kv.val = val
	} else {
		ele := c.ll.PushFront(&entry{key, val})
		c.cache[key] = ele
		c.nBytes += int64(len(key)) + int64(val.Len())
	}
	// remove
	for c.maxBytes != 0 && c.maxBytes < c.nBytes {
		c.RemoveOldest()
	}
}

func (c *lruCache) Get(key string) (val Value, ok bool) {
	if ele, ok := c.cache[key]; ok {
		c.ll.MoveToFront(ele)
		kv := ele.Value.(*entry)
		return kv.val, true
	}
	return
}

func (c *lruCache) RemoveOldest() {
	ele := c.ll.Back()
	if ele != nil {
		c.ll.Remove(ele)
		kv := ele.Value.(*entry)
		delete(c.cache, kv.key)
		c.nBytes -= int64(len(kv.key)) + int64(kv.val.Len())
		if c.OnEvicted != nil {
			c.OnEvicted(kv.key, kv.val)
		}
	}
}

func (c *lruCache) Len() int {
	return c.ll.Len()
}
