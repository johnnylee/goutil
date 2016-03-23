package httpmiddleware

import (
	"compress/gzip"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/johnnylee/goutil/logutil"
)

// We log HTTP requests using the httpLogger.
var httpLogger = logutil.New("http")

// This is a wrapper for the standard http handler that adds request logging,
// including timing.
func Log(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tStart := time.Now().UnixNano()
		handler.ServeHTTP(w, r)
		tEnd := time.Now().UnixNano()
		httpLogger.Msg("%s %s %s %s (%v ms)",
			r.RemoteAddr, r.Method, r.URL, r.Referer(), (tEnd-tStart)/1000000)
	})
}

type GzResponseWriter struct {
	io.Writer
	http.ResponseWriter
}

func (w GzResponseWriter) Write(b []byte) (int, error) {
	// If no content type, apply sniffing algorithm to un-gzipped body.
	if "" == w.Header().Get("Content-Type") {
		w.Header().Set("Content-Type", http.DetectContentType(b))
	}
	return w.Writer.Write(b)
}

// Create a Pool that contains previously used Writers and
// can create new ones if we run out.
var zippers = sync.Pool{New: func() interface{} {
	w, _ := gzip.NewWriterLevel(nil, 3)
	return w
}}

// This wrapper provides gzip compression of outgoing data.
func Gzip(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// If the requester doesn't accept gzip, serve normally.
		if !strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
			handler.ServeHTTP(w, r)
			return
		}

		// Create a gzip writer.
		gz := zippers.Get().(*gzip.Writer)
		defer zippers.Put(gz)
		defer gz.Close()

		gz.Reset(w)

		gw := GzResponseWriter{gz, w}

		// Set headers.
		gw.Header().Set("Content-Encoding", "gzip")
		gw.Header().Set("Vary", "Accept-Encoding")
		handler.ServeHTTP(gw, r)
	})
}
