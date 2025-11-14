package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"time"
)

type ExperimentPayload struct {
	End        int    `json:"end"`
	Current    int    `json:"current"`
	MagnetID   int    `json:"magnet_id"`
	MagnetName string `json:"magnet_name"`
}

type CakeResult struct {
	Experiment string    `json:"experiment"`
	Timestamp  time.Time `json:"timestamp"`
	End        int       `json:"end"`
	Current    int       `json:"current"`
	MagnetID   int       `json:"magnet_id"`
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
			End:        payload.End,
			Current:    payload.Current,
			MagnetID:   payload.MagnetID,
			MagnetName: payload.MagnetName,
			Message:    "All kubrons successfully collided 🎂",
		}

		log.Printf("[%s] detected event from %s (id=%d) end=%d current=%d",
			experimentName, payload.MagnetName, payload.MagnetID, payload.End, payload.Current)

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
