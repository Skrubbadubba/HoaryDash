package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
)

func translationsHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		params := r.URL.Query()

		key := params.Get("key")
		lang := params.Get("lang")
		domain := params.Get("domain")

		if lang == "" {
			http.Error(w, "'lang' parameter must be present", http.StatusBadRequest)
		}

		if domain == "" && key == "" {
			http.Error(w, "no domain or key specified", http.StatusBadRequest)
		}

		var dTranslations map[string]string
		if domain != "" {
			var err error
			dTranslations, err = domainTranslations(domain, lang)
			if err != nil {
				http.Error(w, fmt.Sprintf("domain '%s' not found for language '%s'", domain, lang), http.StatusNotFound)
				return
			}
		}

		if key != "" {
			t, err := translate(fmt.Sprintf("%s:%s", domain, key), lang)
			if err != nil {
				http.Error(w, fmt.Sprintf("key '%s' in domain '%s' not found for language '%s'", key, domain, lang), http.StatusNotFound)
				return
			}
			w.Header().Set("Cache-Control", "max-age=86400")
			json.NewEncoder(w).Encode(map[string]string{"t": t})
		}

		w.Header().Set("Cache-Control", "max-age=86400")
		json.NewEncoder(w).Encode(dTranslations)
	}
}

func mediaCoverHandler(haBaseURL, haToken string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		haBaseURL, haToken = getHaDefaults(haBaseURL, haToken)
		if haBaseURL == "" || haToken == "" {
			http.Error(w, "HA credentials not configured", http.StatusInternalServerError)
			return
		}

		picPath := r.URL.Query().Get("path")
		if picPath == "" {
			http.Error(w, "missing path", http.StatusBadRequest)
			return
		}

		target := strings.TrimRight(haBaseURL, "/") + picPath

		req, err := http.NewRequestWithContext(r.Context(), http.MethodGet, target, nil)
		if err != nil {
			http.Error(w, "bad upstream URL", http.StatusBadRequest)
			return
		}
		req.Header.Set("Authorization", "Bearer "+haToken)

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			log.Printf("media cover upstream error: %v", err)
			http.Error(w, "upstream error", http.StatusBadGateway)
			return
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			http.Error(w, "upstream returned "+resp.Status, resp.StatusCode)
			return
		}

		w.Header().Set("Content-Type", resp.Header.Get("Content-Type"))
		w.Header().Set("Cache-Control", "max-age=300")
		io.Copy(w, resp.Body)
	}
}
