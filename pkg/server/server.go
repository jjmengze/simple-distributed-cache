package server

import (
	"context"
	"fmt"
	"k8s.io/klog/v2"
	"net"
	"net/http"
	"simple-distributed-cache/pkg/cache"
	"simple-distributed-cache/pkg/cache/consistent"
	"time"
)

type Server struct {
	//addr    string
	Server         http.Server
	cache          cache.SetterGetter
	peersNode      *consistent.ConsistentMap
	peerMapHandler map[string]PeerGetter
}

func NewServer(addr string, cache cache.SetterGetter) *Server {
	s := &Server{}
	s.Server.Addr = addr
	s.cache = cache
	http.HandleFunc("/set", s.setHandler)
	http.HandleFunc("/get", s.getHandler)
	return s
}

var db = map[string]string{
	"apple":  "red",
	"banana": "yellow",
	"cat":    "black",
}

//todo fix mock select from db
func mockSelectFromDB(key string) ([]byte, error) {
	klog.V(4).InfoS("[MOCK] search from DB with:", "key", key)
	if v, ok := db[key]; ok {
		return []byte(v), nil
	}
	return nil, fmt.Errorf("%s not exist", key)
}

func (p *Server) Set(hashFn consistent.HashFn, peers ...string) {
	if p.peerMapHandler == nil {
		p.peerMapHandler = make(map[string]PeerGetter)
	}
	p.peersNode = consistent.NewConsistentMAP(consistent.DefaultReplicas, hashFn)
	p.peersNode.Add(peers...)
	for i := 0; i < len(peers); i++ {
		p.peerMapHandler[peers[i]] = NewPeer(peers[i], GetterFunc(mockSelectFromDB), p.cache)
	}
}

//func RunServer(ctx context.Context, server *server.Server, ln net.Listener, shutDownTimeout time.Duration) (context.Context, error) {
func RunServer(ctx context.Context, server *Server, ln net.Listener, shutDownTimeout time.Duration) (<-chan struct{}, error) {
	if ln == nil {
		return nil, fmt.Errorf("listener must not be nil")
	}
	// Shutdown server gracefully.
	stoppedCh := make(chan struct{})
	//serverCtx := context.Background()
	go func() {
		//defer serverCtx.Done()
		//<-ctx.Done()
		defer close(stoppedCh)
		<-ctx.Done()
		//ctx, cancel := context.WithTimeout(serverCtx, shutDownTimeout)
		ctx, cancel := context.WithTimeout(context.Background(), shutDownTimeout)
		err := server.Server.Shutdown(ctx)
		klog.Errorf("failed to shutdown server gracefully: %v", err)
		cancel()
	}()

	go func() {
		//var listener net.Listener
		//listener = tcpKeepAliveListener{ln}
		//if server.TLSConfig != nil {
		//	listener = tls.NewListener(listener, server.TLSConfig)
		//}
		err := server.Server.Serve(ln)

		msg := fmt.Sprintf("Stopped listening on %s", ln.Addr().String())
		select {
		case <-ctx.Done():
			klog.Info(msg)
		default:
			panic(fmt.Sprintf("%s due to error: %v", msg, err))
		}
	}()

	//return serverCtx, nil
	return stoppedCh, nil
}

func (s *Server) setHandler(w http.ResponseWriter, r *http.Request) {
	klog.Infof("[Server %s] %s", r.Method, r.URL.Path)
	key := r.URL.Query().Get("key")
	value := r.URL.Query().Get("value")

	peerNodeIndex := s.peersNode.Get(key)
	handler := s.peerMapHandler[peerNodeIndex].Set(key, value)
	klog.Infof("proxy set cache action to %s cache server", peerNodeIndex)
	handler(w, r)

	//err := s.cache.Set(key, value)
	//if err != nil {
	//	klog.Errorf("set cache key: %v  value :  %v ,error: ", key, value, err)
	//	http.Error(w, err.Error(), http.StatusInternalServerError)
	//	return
	//}
	//w.Header().Set("Content-Type", "application/jason")
	//w.WriteHeader(200)
}

func (s *Server) getHandler(w http.ResponseWriter, r *http.Request) {
	klog.Infof("[Server %s] %s", r.Method, r.URL.Path)
	key := r.URL.Query().Get("key")

	peerNodeIndex := s.peersNode.Get(key)
	handler := s.peerMapHandler[peerNodeIndex].Get(key)
	handler(w, r)

	//val, err := s.cache.Get(key)
	//if err != nil {
	//	klog.Errorf("get cache key: %v  value :  %v ,error: ", key, val, err)
	//	http.Error(w, err.Error(), http.StatusInternalServerError)
	//	return
	//}
	//w.Header().Set("Content-Type", "application/jason")
	//w.WriteHeader(200)
	//klog.Info(val)
	//w.Write([]byte(val))
}
