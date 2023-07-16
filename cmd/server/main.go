package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"

	"github.com/motoki317/swr-cache"
)

var (
	ttl   = flag.String("ttl", "60s", "TTL of cache")
	grace = flag.String("grace", "60s", "Grace period of cache")
	port  = flag.Int("port", 8080, "Listen port of server")
)

func main() {
	flag.Parse()
	handler, err := swrcache.New(&swrcache.Config{
		TTL:   *ttl,
		Grace: *grace,
	})
	if err != nil {
		panic(err)
	}
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", *port), handler))
}
