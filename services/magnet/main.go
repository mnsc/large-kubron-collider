package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

type HopPayload struct {
	End     int `json:"end"`
	Current int `json:"current"`
}

// What magnets send to the experiment
type ExperimentPayload struct {
	End        int    `json:"end"`
	Current    int    `json:"current"`
	MagnetID   int    `json:"magnet_id"`
	MagnetName string `json:"magnet_name"`
}

var (
	myID         int
	ringSize     int
	baseName     string
	serviceDNS   string
	experimentURL string
	httpClient   *http.Client
)

func main() {
	// Config from env
	ringSize = mustGetIntEnv("RING_SIZE")                       // e.g. "1000"
	baseName = getEnv("RING_BASENAME", "magnet")                // StatefulSet name
	serviceDNS = getEnv("RING_SERVICE", "magnets")              // headless Service name
	experimentURL = getEnv("EXPERIMENT_URL", "")                // e.g. "http://experiment-cake:8080/observe"
	addr := getEnv("HTTP_ADDR", ":8080")

	myID = detectMyID(baseName)
	log.Printf("magnet-%d starting, ringSize=%d, experimentURL=%q", myID, ringSize, experimentURL)

	httpClient = &http.Client{
		Timeout: 5 * time.Second,
	}

	mux := http.NewServeMux()
	mux.HandleFunc("POST /hop", hopHandler)

	server := &http.Server{
		Addr:         addr,
		Handler:      mux,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	log.Printf("listening on %s", addr)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("server error: %v", err)
	}
}

func hopHandler(w http.ResponseWriter, r *http.Request) {
	var payload HopPayload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, "invalid JSON body", http.StatusBadRequest)
		return
	}
	if payload.End < 0 {
		http.Error(w, "'end' must be >= 0", http.StatusBadRequest)
		return
	}

	// Done? send to experiment
	if payload.Current >= payload.End {
		if experimentURL == "" {
			http.Error(w, "experiment URL not configured", http.StatusInternalServerError)
			return
		}

		exp := ExperimentPayload{
			End:        payload.End,
			Current:    payload.Current,
			MagnetID:   myID,
			MagnetName: fmt.Sprintf("%s-%d", baseName, myID),
		}

		buf, err := json.Marshal(exp)
		if err != nil {
			http.Error(w, "failed to marshal experiment payload", http.StatusInternalServerError)
			return
		}

		ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
		defer cancel()

		req, err := http.NewRequestWithContext(ctx, http.MethodPost, experimentURL, bytes.NewReader(buf))
		if err != nil {
			http.Error(w, "failed to create experiment request", http.StatusInternalServerError)
			return
		}
		req.Header.Set("Content-Type", "application/json")

		resp, err := httpClient.Do(req)
		if err != nil {
			http.Error(w, "experiment request failed: "+err.Error(), http.StatusBadGateway)
			return
		}
		defer resp.Body.Close()

		// proxy experiment response back
		for k, vals := range resp.Header {
			for _, v := range vals {
				w.Header().Add(k, v)
			}
		}
		w.WriteHeader(resp.StatusCode)
		if _, err := io.Copy(w, resp.Body); err != nil {
			log.Printf("failed to copy experiment response body: %v", err)
		}
		return
	}

	// Not done yet → forward to next magnet
	nextID := (myID + 1) % ringSize
	nextURL := fmt.Sprintf("http://%s-%d.%s:8080/hop", baseName, nextID, serviceDNS)

	nextPayload := HopPayload{
		End:     payload.End,
		Current: payload.Current + 1,
	}

	buf, err := json.Marshal(nextPayload)
	if err != nil {
		http.Error(w, "failed to marshal next payload", http.StatusInternalServerError)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, nextURL, bytes.NewReader(buf))
	if err != nil {
		http.Error(w, "failed to create next magnet request", http.StatusInternalServerError)
		return
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := httpClient.Do(req)
	if err != nil {
		http.Error(w, "next magnet request failed: "+err.Error(), http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()

	for k, vals := range resp.Header {
		for _, v := range vals {
			w.Header().Add(k, v)
		}
	}
	w.WriteHeader(resp.StatusCode)
	if _, err := io.Copy(w, resp.Body); err != nil {
		log.Printf("failed to copy next magnet response body: %v", err)
	}
}

// detectMyID parses hostname like "magnet-12" -> 12
func detectMyID(base string) int {
	hostname, err := os.Hostname()
	if err != nil {
		log.Fatalf("cannot get hostname: %v", err)
	}
	parts := strings.Split(hostname, "-")
	last := parts[len(parts)-1]
	id, err := strconv.Atoi(last)
	if err != nil {
		log.Fatalf("cannot parse ordinal from hostname %q (last part %q): %v", hostname, last, err)
	}
	return id
}

func getEnv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func mustGetIntEnv(key string) int {
	v := os.Getenv(key)
	if v == "" {
		log.Fatalf("missing required env %s", key)
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		log.Fatalf("invalid int value for %s: %q", key, v)
	}
	return n
}
