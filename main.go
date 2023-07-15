package traefikswr

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/motoki317/sc"
)

type Config struct {
	TTL   string `json:"ttl"`
	Grace string `json:"grace"`
}

func CreateConfig() *Config {
	return &Config{}
}

type cacheKey struct {
	method string
	url    string
}

type plugin struct {
	name string
	next http.Handler

	cache *sc.Cache[cacheKey, *response]
}

func New(_ context.Context, next http.Handler, config *Config, name string) (http.Handler, error) {
	ttl, err := time.ParseDuration(config.TTL)
	if err != nil {
		return nil, fmt.Errorf("invalid ttl: %w", err)
	}
	if ttl <= 0 {
		return nil, fmt.Errorf("ttl needs to be a positive value")
	}
	grace, err := time.ParseDuration(config.Grace)
	if err != nil {
		return nil, fmt.Errorf("invalid grace: %w", err)
	}

	p := &plugin{
		name: name,
		next: next,
	}
	p.cache, err = sc.New(p.replace, ttl, ttl+grace, sc.WithCleanupInterval(ttl), sc.EnableStrictCoalescing())
	if err != nil {
		return nil, fmt.Errorf("failed to generate cache instance: %w", err)
	}
	return p, nil
}

var cacheableMethods = []string{
	"HEAD",
	"GET",
}

func (p *plugin) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	if !contains(cacheableMethods, req.Method) {
		p.next.ServeHTTP(rw, req)
		return
	}

	res, err := p.cache.Get(req.Context(), cacheKey{
		method: req.Method,
		url:    req.URL.String(),
	})
	if err != nil {
		log.Printf("error on making request to upstream: %v\n", err)
		return
	}

	header := rw.Header()
	for k, v := range res.headers {
		header[k] = v
	}
	rw.WriteHeader(res.status)
	_, err = rw.Write(res.body.Bytes())
	if err != nil {
		log.Printf("error on writing response: %v\n", err)
		return
	}
}

func (p *plugin) replace(ctx context.Context, key cacheKey) (*response, error) {
	res := newResponse()
	req, err := http.NewRequestWithContext(ctx, key.method, key.url, nil)
	if err != nil {
		return nil, err
	}
	p.next.ServeHTTP(res, req)
	return res, nil
}
