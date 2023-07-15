package traefikswr

import (
	"context"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDemo(t *testing.T) {
	cfg := CreateConfig()
	cfg.TTL = "60s"
	cfg.Grace = "60s"

	// Prepare
	ctx := context.Background()
	var called int64
	next := http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		rw.Header()["Test"] = []string{"value"}
		rw.WriteHeader(200)
		_, err := rw.Write([]byte(`test body`))
		require.NoError(t, err)
		atomic.AddInt64(&called, 1)
	})

	handler, err := New(ctx, next, cfg, "swr")
	require.NoError(t, err)

	// Test
	runAndValidate := func() {
		recorder := httptest.NewRecorder()
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, "http://localhost", nil)
		require.NoError(t, err)
		handler.ServeHTTP(recorder, req)

		// Validate
		headers := recorder.Header()["Test"]
		if assert.Len(t, headers, 1) {
			assert.EqualValues(t, "value", headers[0])
		}
		assert.EqualValues(t, 200, recorder.Code)
		assert.EqualValues(t, `test body`, recorder.Body.String())
	}

	assert.EqualValues(t, 0, atomic.LoadInt64(&called))
	for i := 0; i < 10; i++ {
		runAndValidate()
		assert.EqualValues(t, 1, atomic.LoadInt64(&called))
	}
}
