package leads

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"strings"
	"time"

	"portfolio/internal/database"
	"portfolio/internal/email"

	"github.com/google/uuid"
)

type Lead struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Email       string    `json:"email"`
	Phone       string    `json:"phone"`
	Company     string    `json:"company"`
	ProjectType string    `json:"project_type"`
	Budget      string    `json:"budget"`
	Timeline    string    `json:"timeline"`
	Message     string    `json:"message"`
	Status      string    `json:"status"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type LeadNote struct {
	ID        string    `json:"id"`
	LeadID    string    `json:"lead_id"`
	Note      string    `json:"note"`
	CreatedAt time.Time `json:"created_at"`
}

func SubmitHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var lead Lead
	if err := json.NewDecoder(r.Body).Decode(&lead); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if strings.TrimSpace(lead.Name) == "" || strings.TrimSpace(lead.Email) == "" {
		http.Error(w, "Name and email are required", http.StatusBadRequest)
		return
	}

	lead.ID = uuid.New().String()
	lead.Status = "new"

	_, err := database.DB.Exec(`
		INSERT INTO leads (id, name, email, phone, company, project_type, budget, timeline, message, status)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)`,
		lead.ID, lead.Name, lead.Email, lead.Phone, lead.Company,
		lead.ProjectType, lead.Budget, lead.Timeline, lead.Message, lead.Status,
	)
	if err != nil {
		http.Error(w, "Failed to save lead", http.StatusInternalServerError)
		return
	}

	// The lead is safely stored, so the visitor gets their confirmation whether or
	// not the notification email succeeds. Detached from r.Context(), which is
	// canceled as soon as the response is written.
	go func(l Lead) {
		ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
		defer cancel()

		if err := email.NotifyNewLead(ctx, email.LeadNotification{
			Name:        l.Name,
			Email:       l.Email,
			Phone:       l.Phone,
			Company:     l.Company,
			ProjectType: l.ProjectType,
			Budget:      l.Budget,
			Timeline:    l.Timeline,
			Message:     l.Message,
		}); err != nil {
			log.Printf("[leads] notification failed for %s: %v", l.ID, err)
		}
	}(lead)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{"id": lead.ID, "status": "received"})
}

func ListHandler(w http.ResponseWriter, r *http.Request) {
	status := r.URL.Query().Get("status")
	query := `SELECT id, name, email, phone, company, project_type, budget, timeline, message, status, created_at, updated_at FROM leads`
	args := []interface{}{}
	if status != "" {
		query += " WHERE status = $1"
		args = append(args, status)
	}
	query += " ORDER BY created_at DESC"

	rows, err := database.DB.Query(query, args...)
	if err != nil {
		http.Error(w, "Failed to fetch leads", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var leads []Lead
	for rows.Next() {
		var l Lead
		rows.Scan(&l.ID, &l.Name, &l.Email, &l.Phone, &l.Company,
			&l.ProjectType, &l.Budget, &l.Timeline, &l.Message, &l.Status, &l.CreatedAt, &l.UpdatedAt)
		leads = append(leads, l)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(leads)
}

func UpdateStatusHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	parts := strings.Split(r.URL.Path, "/")
	id := parts[len(parts)-2]

	var body struct {
		Status string `json:"status"`
	}
	json.NewDecoder(r.Body).Decode(&body)

	validStatuses := map[string]bool{
		"new": true, "contacted": true, "meeting_scheduled": true,
		"proposal_sent": true, "negotiation": true, "won": true, "lost": true,
	}
	if !validStatuses[body.Status] {
		http.Error(w, "Invalid status", http.StatusBadRequest)
		return
	}

	_, err := database.DB.Exec(
		"UPDATE leads SET status=$1, updated_at=NOW() WHERE id=$2",
		body.Status, id,
	)
	if err != nil {
		http.Error(w, "Failed to update lead", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func AddNoteHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	parts := strings.Split(r.URL.Path, "/")
	leadID := parts[len(parts)-2]

	var body struct {
		Note string `json:"note"`
	}
	json.NewDecoder(r.Body).Decode(&body)

	noteID := uuid.New().String()
	_, err := database.DB.Exec(
		"INSERT INTO lead_notes (id, lead_id, note) VALUES ($1,$2,$3)",
		noteID, leadID, body.Note,
	)
	if err != nil {
		http.Error(w, "Failed to add note", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
}
