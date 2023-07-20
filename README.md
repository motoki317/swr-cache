# swr cache server

A dead-simple swr (stale-while-revalidate) style cache server.

- Coalesces requests and requests the upstream in the background while in "grace" period.
- Dead-simple as in, ignores all cache-control headers and just caches GET / HEAD requests.
No cache conditions supported (yet).

## Testing

- Run the cache server: `go run ./cmd/server`
- Run some service at `localhost:80`: ` docker run --rm -it --network host caddy:latest caddy file-server --root /usr/share/caddy --debug`
