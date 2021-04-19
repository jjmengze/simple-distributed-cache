package server

import (
	"context"
	"fmt"
	"k8s.io/klog/v2"
	"net"
	"net/http"
	"simple-distributed-cache/pkg/cache"
	"time"
)

type Server struct {
	//addr    string
	Server http.Server
	cache  cache.SetterGetter
}

func NewServer(addr string, cache cache.SetterGetter) *Server {
	s := &Server{}
	s.Server.Addr = addr
	s.cache = cache
	http.HandleFunc("/set", s.setHandler)
	http.HandleFunc("/get", s.getHandler)
	return s
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
	err := s.cache.Set(key, value)
	if err != nil {
		klog.Errorf("set cache key: %v  value :  %v ,error: ", key, value, err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/jason")
	w.WriteHeader(200)
}

func (s *Server) getHandler(w http.ResponseWriter, r *http.Request) {
	klog.Infof("[Server %s] %s", r.Method, r.URL.Path)
	key := r.URL.Query().Get("key")
	val, err := s.cache.Get(key)
	if err != nil {
		klog.Errorf("get cache key: %v  value :  %v ,error: ", key, val, err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/jason")
	w.WriteHeader(200)
	klog.Info(val)
	w.Write([]byte(val))
}
