package analytics

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"net/http"

	"portfolio/internal/database"

	"github.com/google/uuid"
)

type Event struct {
	EventType string `json:"event_type"`
	Page      string `json:"page"`
	SessionID string `json:"session_id"`
}

func TrackHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var event Event
	json.NewDecoder(r.Body).Decode(&event)

	ip := r.RemoteAddr
	hash := fmt.Sprintf("%x", sha256.Sum256([]byte(ip)))

	database.DB.Exec(
		`INSERT INTO analytics_events (id, event_type, page, session_id, ip_hash) VALUES ($1,$2,$3,$4,$5)`,
		uuid.New().String(), event.EventType, event.Page, event.SessionID, hash,
	)

	w.WriteHeader(http.StatusNoContent)
}

type PageStat struct {
	Page  string `json:"page"`
	Views int    `json:"views"`
}

func PageStatsHandler(w http.ResponseWriter, r *http.Request) {
	rows, err := database.DB.Query(
		`SELECT page, COUNT(*) as views FROM analytics_events WHERE event_type='pageview' GROUP BY page ORDER BY views DESC LIMIT 20`,
	)
	if err != nil {
		http.Error(w, "Failed to fetch stats", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var stats []PageStat
	for rows.Next() {
		var s PageStat
		rows.Scan(&s.Page, &s.Views)
		stats = append(stats, s)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(stats)
}
