package proxy

import (
	"context"
	"log/slog"
	"net/http"
	"net/http/httputil"
	"sync"

	"lb/internal/backend"
)

type backendKey struct{}

type bytePool struct {
	pool sync.Pool
}

func newBytePool() *bytePool {
	return &bytePool{
		pool: sync.Pool{
			New: func() any {
				b := make([]byte, 32*1024)
				return &b
			},
		},
	}
}

func (p *bytePool) Get() []byte {
	return *p.pool.Get().(*[]byte)
}

func (p *bytePool) Put(b []byte) {
	p.pool.Put(&b)
}

type Proxy struct {
	proxy *httputil.ReverseProxy
}

func NewProxy(t http.RoundTripper) *Proxy {
	bp := newBytePool()

	p := &httputil.ReverseProxy{
		Transport:  t,
		BufferPool: bp,
		Rewrite: func(pr *httputil.ProxyRequest) {
			val := pr.In.Context().Value(backendKey{})
			b, ok := val.(*backend.Backend)
			if !ok || b == nil {
				slog.Error("backend missing in request context")
				return
			}

			pr.SetURL(b.URL)
			pr.Out.Host = pr.In.Host
		},
		ErrorHandler: func(w http.ResponseWriter, r *http.Request, err error) {
			slog.Error("proxy error", "error", err)
			http.Error(w, "Bad Gateway", http.StatusBadGateway)
		},
	}

	return &Proxy{proxy: p}
}

func WithBackend(r *http.Request, b *backend.Backend) *http.Request {
	return r.WithContext(context.WithValue(r.Context(), backendKey{}, b))
}

func (p *Proxy) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	p.proxy.ServeHTTP(w, r)
}
