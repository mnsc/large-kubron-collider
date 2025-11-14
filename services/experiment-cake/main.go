package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"
)

type ExperimentPayload struct {
	Current    int    `json:"current"`
	MagnetName string `json:"magnet_name"`
}

type CakeResult struct {
	Experiment string    `json:"experiment"`
	Timestamp  time.Time `json:"timestamp"`
	Current    int       `json:"current"`
	MagnetName string    `json:"magnet_name"`
	Message    string    `json:"message"`
}

func main() {
	addr := getEnv("HTTP_ADDR", ":8080")
	experimentName := getEnv("EXPERIMENT_NAME", "CAKE") // Colliding Automation Kubernetes Experiment

	mux := http.NewServeMux()
	mux.HandleFunc("POST /observe", func(w http.ResponseWriter, r *http.Request) {
		var payload ExperimentPayload
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			http.Error(w, "invalid JSON body", http.StatusBadRequest)
			return
		}

		result := CakeResult{
			Experiment: experimentName,
			Timestamp:  time.Now().UTC(),
			Current:    payload.Current,
			MagnetName: payload.MagnetName,
			Message:    fmt.Sprintf("Kubron of size %d observed 🎂", payload.Current),
		}

		log.Printf("[%s] detected event from %s current=%d",
			experimentName, payload.MagnetName, payload.Current)

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(result)
	})

	log.Printf("starting experiment %s on %s", experimentName, addr)
	if err := http.ListenAndServe(addr, mux); err != nil && err != http.ErrServerClosed {
		log.Fatalf("server error: %v", err)
	}
}

func getEnv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
