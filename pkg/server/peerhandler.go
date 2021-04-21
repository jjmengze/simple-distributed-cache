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

//type PeerGetter func()

type Peer struct {
	name  string
	cache cache.SetterGetter
}

func NewPeer(name string, cache cache.SetterGetter) PeerGetter {
	return &Peer{
		name:  name,
		cache: cache,
	}
}

func (p *Peer) Get(key string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		val, err := p.cache.Get(key)
		if err != nil {
			klog.Errorf("peer node get cache key: %v  value :  %v ,error: ", key, val, err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/jason")
		w.WriteHeader(200)
		klog.Info(val)
		w.Write([]byte(val))
	}
}
func (p *Peer) Set(key string, value interface{}) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		err := p.cache.Set(key, value.(string))
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
