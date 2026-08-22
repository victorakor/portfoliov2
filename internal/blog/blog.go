package blog

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"portfolio/internal/database"

	"github.com/google/uuid"
)

type Post struct {
	ID         string    `json:"id"`
	Title      string    `json:"title"`
	Slug       string    `json:"slug"`
	Excerpt    string    `json:"excerpt"`
	Content    string    `json:"content"`
	CoverImage string    `json:"cover_image"`
	Category   string    `json:"category"`
	Published  bool      `json:"published"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

func ListPublicHandler(w http.ResponseWriter, r *http.Request) {
	category := r.URL.Query().Get("category")
	// excerpt/cover_image/category are nullable and rows.Scan errors are ignored,
	// so COALESCE stops one NULL from blanking a whole row.
	query := `SELECT id, title, slug, COALESCE(excerpt,''), COALESCE(cover_image,''),
	                 COALESCE(category,''), published, created_at, updated_at
	           FROM blog_posts WHERE published = true`
	args := []interface{}{}
	if category != "" {
		query += " AND category = $1"
		args = append(args, category)
	}
	query += " ORDER BY created_at DESC"

	rows, err := database.DB.Query(query, args...)
	if err != nil {
		http.Error(w, "Failed to fetch posts", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var posts []Post
	for rows.Next() {
		var p Post
		rows.Scan(&p.ID, &p.Title, &p.Slug, &p.Excerpt, &p.CoverImage, &p.Category, &p.Published, &p.CreatedAt, &p.UpdatedAt)
		posts = append(posts, p)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(posts)
}

func GetBySlugHandler(w http.ResponseWriter, r *http.Request) {
	parts := strings.Split(r.URL.Path, "/")
	slug := parts[len(parts)-1]

	var p Post
	err := database.DB.QueryRow(
		`SELECT id, title, slug, COALESCE(excerpt,''), COALESCE(content,''), COALESCE(cover_image,''),
		        COALESCE(category,''), published, created_at, updated_at
		 FROM blog_posts WHERE slug=$1 AND published=true`, slug,
	).Scan(&p.ID, &p.Title, &p.Slug, &p.Excerpt, &p.Content, &p.CoverImage, &p.Category, &p.Published, &p.CreatedAt, &p.UpdatedAt)
	if err != nil {
		http.Error(w, "Post not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(p)
}

func AdminListHandler(w http.ResponseWriter, r *http.Request) {
	// content is included here (unlike the public list) because the admin edit
	// form prefills from this payload — without it, saving would blank the body.
	rows, err := database.DB.Query(
		`SELECT id, title, slug, COALESCE(excerpt,''), COALESCE(content,''), COALESCE(cover_image,''),
		        COALESCE(category,''), published, created_at, updated_at
		 FROM blog_posts ORDER BY created_at DESC`,
	)
	if err != nil {
		http.Error(w, "Failed to fetch posts", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var posts []Post
	for rows.Next() {
		var p Post
		rows.Scan(&p.ID, &p.Title, &p.Slug, &p.Excerpt, &p.Content, &p.CoverImage, &p.Category, &p.Published, &p.CreatedAt, &p.UpdatedAt)
		posts = append(posts, p)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(posts)
}

func CreateHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var p Post
	json.NewDecoder(r.Body).Decode(&p)
	p.ID = uuid.New().String()
	p.Slug = slugify(p.Title)

	_, err := database.DB.Exec(
		`INSERT INTO blog_posts (id, title, slug, excerpt, content, cover_image, category, published)
		 VALUES ($1,$2,$3,$4,$5,$6,$7,$8)`,
		p.ID, p.Title, p.Slug, p.Excerpt, p.Content, p.CoverImage, p.Category, p.Published,
	)
	if err != nil {
		http.Error(w, "Failed to create post", http.StatusInternalServerError)
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

	var p Post
	json.NewDecoder(r.Body).Decode(&p)

	_, err := database.DB.Exec(
		`UPDATE blog_posts SET title=$1, excerpt=$2, content=$3, cover_image=$4, category=$5, published=$6, updated_at=NOW() WHERE id=$7`,
		p.Title, p.Excerpt, p.Content, p.CoverImage, p.Category, p.Published, id,
	)
	if err != nil {
		http.Error(w, "Failed to update post", http.StatusInternalServerError)
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

	database.DB.Exec("DELETE FROM blog_posts WHERE id=$1", id)
	w.WriteHeader(http.StatusNoContent)
}

func slugify(s string) string {
	s = strings.ToLower(s)
	s = strings.ReplaceAll(s, " ", "-")
	return s
}
