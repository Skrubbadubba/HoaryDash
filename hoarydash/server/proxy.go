package main

import (
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"
)

func haProxyHandler(haUrl string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		path := strings.TrimPrefix(r.URL.Path, "/api/proxy")
		target, err := url.Parse(haUrl + path)
		if err != nil {
			http.Error(w, "bad gateway", http.StatusBadGateway)
			return
		}
		target.RawQuery = r.URL.RawQuery

		req, err := http.NewRequestWithContext(r.Context(), r.Method, target.String(), r.Body)
		if err != nil {
			http.Error(w, "bad gateway", http.StatusBadGateway)
			return
		}

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			log.Printf("HA proxy error: %v", err)
			http.Error(w, "bad gateway", http.StatusBadGateway)
			return
		}
		defer resp.Body.Close()

		for key, vals := range resp.Header {
			for _, v := range vals {
				w.Header().Add(key, v)
			}
		}
		w.WriteHeader(resp.StatusCode)
		io.Copy(w, resp.Body)
	}
}
