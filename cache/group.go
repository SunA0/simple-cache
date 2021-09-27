package cache

import (
	"fmt"
	"log"
	"simpleCache/singleflight"
	"sync"
)

var (
	groupLock sync.RWMutex
	groups    = make(map[string]*Group)
)

type Group struct {
	name      string
	mainCache cache
	getter    Getter

	peers PeerPicker

	loader *singleflight.Group
}

func NewGroup(name string, cacheBytes int64, getter Getter) *Group {
	if getter == nil {
		panic("nil Getter")
	}
	groupLock.Lock()
	defer groupLock.Unlock()
	g := &Group{
		name:      name,
		getter:    getter,
		mainCache: cache{cacheBytes: cacheBytes},

		loader: &singleflight.Group{},
	}
	groups[name] = g
	return g
}

func NewGroupFunc(name string, cacheBytes int64, getter GetterFunc) *Group {
	return NewGroup(name, cacheBytes, GetterFunc(getter))
}

func GetGroup(name string) *Group {
	groupLock.RLock()
	g := groups[name]
	groupLock.RUnlock()
	return g
}

func (g *Group) RegisterPeers(peers PeerPicker) {
	if g.peers != nil {
		panic("RegisterPeerPicker called more than once")
	}
	g.peers = peers
}

func (g *Group) Get(key string) (ByteView, error) {
	// no key
	if key == "" {
		return ByteView{}, fmt.Errorf("key is required")
	}
	// find cache
	if v, ok := g.mainCache.Get(key); ok {
		log.Println("[Cache] hit", key, v)
		return v, nil
	}
	// not found and need load
	return g.load(key)
}

// load
// return getLocally
func (g *Group) load(key string) (value ByteView, err error) {
	// each key is only fetched once (either locally or remotely)
	// regardless of the number of concurrent callers.
	viewi, err := g.loader.Do(key, func() (interface{}, error) {
		if g.peers != nil {
			if peer, ok := g.peers.PickPeer(key); ok {
				if value, err = g.getFromPeer(peer, key); err == nil {
					return value, nil
				}
				log.Println("[Cache] Failed to get from peer", err)
			}
		}

		return g.getLocally(key)
	})

	if err == nil {
		return viewi.(ByteView), nil
	}
	return
}

func (g *Group) getFromPeer(peer PeerGetter, key string) (ByteView, error) {
	bytes, err := peer.Get(g.name, key)
	if err != nil {
		return ByteView{}, err
	}
	return ByteView{b: bytes}, nil
}

// getLocally
// 添加缓存值
func (g *Group) getLocally(key string) (ByteView, error) {
	bytes, err := g.getter.Get(key)
	if err != nil {
		return ByteView{}, err
	}
	value := ByteView{b: CloneBytes(bytes)}
	g.populateCache(key, value)
	return value, nil
}

func (g *Group) populateCache(key string, value ByteView) {
	g.mainCache.Add(key, value)
}
