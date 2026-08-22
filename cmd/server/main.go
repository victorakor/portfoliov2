package main

import (
	"log"
	"net/http"
	"os"
	"sort"
	"strings"

	"portfolio/internal/analytics"
	"portfolio/internal/assistant"
	"portfolio/internal/auth"
	"portfolio/internal/blog"
	"portfolio/internal/dashboard"
	"portfolio/internal/database"
	"portfolio/internal/leads"
	"portfolio/internal/middleware"
	"portfolio/internal/projects"
	"portfolio/internal/storage"

	"github.com/joho/godotenv"
)

func main() {
	_ = godotenv.Load()

	database.Connect()

	mux := http.NewServeMux()

	// Static files
	mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

	// Public pages
	mux.HandleFunc("/", serveTemplate("templates/public/index.html"))
	mux.HandleFunc("/about", serveTemplate("templates/public/about.html"))
	mux.HandleFunc("/blog", serveTemplate("templates/public/blog.html"))
	mux.HandleFunc("/blog/", serveTemplate("templates/public/blog-post.html"))
	mux.HandleFunc("/contact", serveTemplate("templates/public/contact.html"))

	// Case study pages
	mux.HandleFunc("/projects/mall-surveillance-system", serveTemplate("templates/public/projects/mall-surveillance.html"))
	mux.HandleFunc("/projects/face-recognition-system", serveTemplate("templates/public/projects/face-recognition.html"))
	mux.HandleFunc("/projects/eye-disease-detection", serveTemplate("templates/public/projects/eye-disease.html"))
	mux.HandleFunc("/projects/hackerthon", serveTemplate("templates/public/projects/hackerthon.html"))
	mux.HandleFunc("/projects/text-analyzer", serveTemplate("templates/public/projects/text-analyzer.html"))
	mux.HandleFunc("/projects/cli-calculator", serveTemplate("templates/public/projects/cli-calculator.html"))
	mux.HandleFunc("/projects/gwinks-hub", serveTemplate("templates/public/projects/gwinks-hub.html"))

	// Public API
	mux.HandleFunc("/api/contact", leads.SubmitHandler)
	mux.HandleFunc("/api/blog", blog.ListPublicHandler)
	mux.HandleFunc("/api/blog/", blog.GetBySlugHandler)
	mux.HandleFunc("/api/projects", projects.ListHandler)
	mux.HandleFunc("/api/track", analytics.TrackHandler)
	mux.HandleFunc("/api/assistant/chat", assistant.ChatHandler)

	// Admin auth
	mux.HandleFunc("/admin/login", auth.LoginHandler)
	mux.HandleFunc("/admin/logout", auth.LogoutHandler)

	// Admin pages (protected)
	adminMux := http.NewServeMux()
	adminMux.HandleFunc("/admin", serveTemplate("templates/admin/dashboard.html"))
	adminMux.HandleFunc("/admin/leads", serveTemplate("templates/admin/leads.html"))
	adminMux.HandleFunc("/admin/leads/", serveTemplate("templates/admin/lead-detail.html"))
	adminMux.HandleFunc("/admin/email", serveTemplate("templates/admin/email-center.html"))
	adminMux.HandleFunc("/admin/blog", serveTemplate("templates/admin/blog-manage.html"))
	adminMux.HandleFunc("/admin/projects", serveTemplate("templates/admin/projects-manage.html"))
	adminMux.HandleFunc("/admin/analytics", serveTemplate("templates/admin/analytics.html"))

	// Admin API (protected)
	adminMux.HandleFunc("/api/admin/stats", dashboard.StatsHandler)
	adminMux.HandleFunc("/api/admin/leads", leads.ListHandler)
	adminMux.HandleFunc("/api/admin/leads/", leads.UpdateStatusHandler)
	adminMux.HandleFunc("/api/admin/blog", blog.AdminListHandler)
	adminMux.HandleFunc("/api/admin/blog/create", blog.CreateHandler)
	adminMux.HandleFunc("/api/admin/blog/", byMethod(map[string]http.HandlerFunc{
		http.MethodPut:    blog.UpdateHandler,
		http.MethodDelete: blog.DeleteHandler,
	}))
	adminMux.HandleFunc("/api/admin/projects", projects.ListHandler)
	adminMux.HandleFunc("/api/admin/projects/create", projects.CreateHandler)
	adminMux.HandleFunc("/api/admin/projects/", byMethod(map[string]http.HandlerFunc{
		http.MethodPut:    projects.UpdateHandler,
		http.MethodDelete: projects.DeleteHandler,
	}))
	adminMux.HandleFunc("/api/admin/upload", storage.UploadHandler)
	adminMux.HandleFunc("/api/admin/analytics/pages", analytics.PageStatsHandler)

	mux.Handle("/admin", middleware.Auth(adminMux))
	mux.Handle("/admin/", middleware.Auth(adminMux))
	mux.Handle("/api/admin/", middleware.Auth(adminMux))

	handler := middleware.SecurityHeaders(middleware.CORS(middleware.Logger(mux)))

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Server starting on :%s", port)
	if err := http.ListenAndServe(":"+port, handler); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}

func serveTemplate(path string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, path)
	}
}

// byMethod dispatches one path to a different handler per HTTP method. Needed
// because ServeMux matches on path only, so /api/admin/blog/ has to serve both
// PUT (update) and DELETE (delete).
func byMethod(handlers map[string]http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		h, ok := handlers[r.Method]
		if !ok {
			allowed := make([]string, 0, len(handlers))
			for m := range handlers {
				allowed = append(allowed, m)
			}
			sort.Strings(allowed)
			w.Header().Set("Allow", strings.Join(allowed, ", "))
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		h(w, r)
	}
}
