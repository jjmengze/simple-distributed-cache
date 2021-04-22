package server

import (
	"k8s.io/klog/v2"
	"net/http"
	"simple-distributed-cache/pkg/cache"
)

type PeerGetter interface {
	Get(key string) http.HandlerFunc
	Set(key string, value interface{}) http.HandlerFunc
}

// A Getter loads data for a key.
type Getter interface {
	Get(key string) (interface{}, error)
}

type GetterFunc func(key string) (interface{}, error)

func (f GetterFunc) Get(key string) (interface{}, error) {
	return f(key)
}

type Peer struct {
	name        string
	dataHandler Getter
	cache       cache.SetterGetter
}

func NewPeer(name string, dataHandler Getter, cache cache.SetterGetter) PeerGetter {
	return &Peer{
		name:        name,
		dataHandler: dataHandler,
		cache:       cache,
	}
}

func (p *Peer) Get(key string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		val, ok := p.cache.Get(key)
		var err error
		if !ok && p.dataHandler != nil {
			klog.V(4).InfoS("peer node not HIT cache, try to get data from remote", "server", p.name, "key", key)
			val, err = p.getDataFromRemote(key)
		}
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
		}

		w.Header().Set("Content-Type", "application/jason")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(val.(string)))
	}
}
func (p *Peer) Set(key string, value interface{}) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		err := p.cache.Set(key, value)
		klog.Infof("Set cache key : %s value : %v", key, value)
		if err != nil {
			klog.Errorf("peer node set cache key: %v  value :  %v ,error: ", key, value, err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/jason")
		w.WriteHeader(200)
	}
}

func (p *Peer) getDataFromRemote(key string) (interface{}, error) {
	b, err := p.dataHandler.Get(key)
	if err != nil {
		klog.Errorf("get remove data fail err %v", err)
	}
	err = p.putDataToCache(key, b)
	if err != nil {
		klog.Errorf("put data to cache fail err %v", err)
	}
	return b, err
}

func (p *Peer) putDataToCache(key string, value interface{}) error {
	return p.cache.Set(key, value)
}
