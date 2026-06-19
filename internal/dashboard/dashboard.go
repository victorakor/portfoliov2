package dashboard

import (
	"encoding/json"
	"net/http"

	"portfolio/internal/database"
)

type Stats struct {
	TotalLeads   int `json:"total_leads"`
	NewLeads     int `json:"new_leads"`
	Contacted    int `json:"contacted"`
	Negotiation  int `json:"negotiation"`
	Won          int `json:"won"`
	Lost         int `json:"lost"`
	TotalPosts   int `json:"total_posts"`
	PublishedPosts int `json:"published_posts"`
	TotalVisitors int `json:"total_visitors"`
	TotalSessions int `json:"total_sessions"`
}

func StatsHandler(w http.ResponseWriter, r *http.Request) {
	var stats Stats

	database.DB.QueryRow(`SELECT COUNT(*) FROM leads`).Scan(&stats.TotalLeads)
	database.DB.QueryRow(`SELECT COUNT(*) FROM leads WHERE status='new'`).Scan(&stats.NewLeads)
	database.DB.QueryRow(`SELECT COUNT(*) FROM leads WHERE status='contacted'`).Scan(&stats.Contacted)
	database.DB.QueryRow(`SELECT COUNT(*) FROM leads WHERE status='negotiation'`).Scan(&stats.Negotiation)
	database.DB.QueryRow(`SELECT COUNT(*) FROM leads WHERE status='won'`).Scan(&stats.Won)
	database.DB.QueryRow(`SELECT COUNT(*) FROM leads WHERE status='lost'`).Scan(&stats.Lost)
	database.DB.QueryRow(`SELECT COUNT(*) FROM blog_posts`).Scan(&stats.TotalPosts)
	database.DB.QueryRow(`SELECT COUNT(*) FROM blog_posts WHERE published=true`).Scan(&stats.PublishedPosts)
	database.DB.QueryRow(`SELECT COUNT(DISTINCT ip_hash) FROM analytics_events WHERE event_type='pageview'`).Scan(&stats.TotalVisitors)
	database.DB.QueryRow(`SELECT COUNT(DISTINCT session_id) FROM analytics_events`).Scan(&stats.TotalSessions)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(stats)
}
