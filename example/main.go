package main

import (
	"flag"
	"log"
	"net/http"
	"simpleCache/cache"
)

var db = map[string]string{
	"a": "123",
	"b": "ccc",
	"c": "ddd",
}

//
//func main() {
//	cache.NewGroup("map", 2<<2, cache.GetterFunc(getter))
//	addr := "localhost:9999"
//	peers := cache.NewHttpPool(addr)
//	log.Println("cache is running at", addr)
//	log.Fatal(http.ListenAndServe(addr, peers))
//}
//
//func getter(key string) ([]byte, error) {
//	log.Println("[search key]", key, "...")
//	if v, ok := db[key]; ok {
//		return []byte(v), nil
//	}
//	return nil, fmt.Errorf("%s not exist", key)
//}

// example2
func main() {
	var port int
	var api bool

	flag.IntVar(&port, "port", 8001, "cache server port")
	flag.BoolVar(&api, "api", false, "Start a api server?")
	flag.Parse()

	apiAddr := "http://localhost:9999"
	addrMap := map[int]string{
		8001: "http://localhost:8001",
		8002: "http://localhost:8002",
		8003: "http://localhost:8003",
	}

	var addrs []string
	for _, v := range addrMap {
		addrs = append(addrs, v)
	}
	//=> ["","",""]

	group := cache.NewGroupFunc(
		"maps",
		2<<10,
		cache.DefaultGetterHandler)
	if api {
		go startAPIServer(apiAddr, group)
	}
	startCacheServer(addrMap[port], []string(addrs), group)
}

func startAPIServer(apiAddr string, group *cache.Group) {
	http.HandleFunc(
		"/api",
		func(w http.ResponseWriter, r *http.Request) {
			key := r.URL.Query().Get("key")
			view, err := group.Get(key)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			w.Header().Set("Content-Type", "application/octet-stream")
			w.Write(view.ByteSlice())
		})
	log.Println("frontend server is running at ", apiAddr)
	log.Fatal(http.ListenAndServe(apiAddr[7:], nil))
}

func startCacheServer(addr string, addrs []string, group *cache.Group) {
	peers := cache.NewHTTPPool(addr)
	peers.Set(addrs...)
	group.RegisterPeers(peers)
	log.Println("cache is running at", addr)
	log.Fatal(http.ListenAndServe(addr[7:], peers))
}
