package projects

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"portfolio/internal/database"

	"github.com/google/uuid"
)

type Project struct {
	ID          string    `json:"id"`
	Title       string    `json:"title"`
	Slug        string    `json:"slug"`
	Description string    `json:"description"`
	GithubURL   string    `json:"github_url"`
	LiveURL     string    `json:"live_url"`
	ImageURL    string    `json:"image_url"`
	Featured    bool      `json:"featured"`
	Archived    bool      `json:"archived"`
	CreatedAt   time.Time `json:"created_at"`
}

func ListHandler(w http.ResponseWriter, r *http.Request) {
	// These TEXT columns are all nullable, and rows.Scan errors are ignored below,
	// so a single NULL would silently blank an entire row. COALESCE keeps that
	// from happening — image_url in particular is NULL on every pre-existing row.
	rows, err := database.DB.Query(
		`SELECT id, title, slug, COALESCE(description,''), COALESCE(github_url,''),
		        COALESCE(live_url,''), COALESCE(image_url,''), featured, archived, created_at
		 FROM projects WHERE archived = false ORDER BY featured DESC, created_at DESC`,
	)
	if err != nil {
		http.Error(w, "Failed to fetch projects", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var projects []Project
	for rows.Next() {
		var p Project
		rows.Scan(&p.ID, &p.Title, &p.Slug, &p.Description, &p.GithubURL, &p.LiveURL, &p.ImageURL, &p.Featured, &p.Archived, &p.CreatedAt)
		projects = append(projects, p)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(projects)
}

func CreateHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var p Project
	json.NewDecoder(r.Body).Decode(&p)
	p.ID = uuid.New().String()
	p.Slug = slugify(p.Title)

	_, err := database.DB.Exec(
		`INSERT INTO projects (id, title, slug, description, github_url, live_url, image_url, featured)
		 VALUES ($1,$2,$3,$4,$5,$6,$7,$8)`,
		p.ID, p.Title, p.Slug, p.Description, p.GithubURL, p.LiveURL, p.ImageURL, p.Featured,
	)
	if err != nil {
		http.Error(w, "Failed to create project", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(p)
}

func UpdateHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	parts := strings.Split(r.URL.Path, "/")
	id := parts[len(parts)-1]

	var p Project
	json.NewDecoder(r.Body).Decode(&p)

	_, err := database.DB.Exec(
		`UPDATE projects SET title=$1, description=$2, github_url=$3, live_url=$4, image_url=$5, featured=$6, archived=$7 WHERE id=$8`,
		p.Title, p.Description, p.GithubURL, p.LiveURL, p.ImageURL, p.Featured, p.Archived, id,
	)
	if err != nil {
		http.Error(w, "Failed to update project", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func DeleteHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	parts := strings.Split(r.URL.Path, "/")
	id := parts[len(parts)-1]

	database.DB.Exec("DELETE FROM projects WHERE id=$1", id)
	w.WriteHeader(http.StatusNoContent)
}

func slugify(s string) string {
	s = strings.ToLower(s)
	s = strings.ReplaceAll(s, " ", "-")
	return s
}
