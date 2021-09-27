package cache

import "sync"

type cache struct {
	mu         sync.Mutex
	lru        *lruCache
	cacheBytes int64
}

func (c *cache) Add(key string, val ByteView) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.lru == nil {
		c.lru = NewLru(c.cacheBytes, nil)
	}
	c.lru.Add(key, val)
}

func (c *cache) Get(key string) (ByteView, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.lru == nil {
		return ByteView{}, false
	}
	if v, ok := c.lru.Get(key); ok {
		return v.(ByteView), ok
	}
	return ByteView{}, false
}
