package assistant

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
)

// Ollama Cloud's OpenAI-compatible chat endpoint.
const ollamaAPIURL = "https://ollama.com/v1/chat/completions"

// defaultModel is the general-purpose Ollama Cloud model. Measured at ~1.7s for
// a 300-token reply, comfortably inside httpClient's 20s timeout. The larger
// gpt-oss:120b answers better than gpt-oss:20b at no meaningful latency cost.
// Override with the OLLAMA_MODEL env var.
const defaultModel = "gpt-oss:120b"

const systemPrompt = `You are the AI assistant embedded on Victor Akor's portfolio website.
Victor is a  Software Engineer & AI Specialist with 5+ years of experience.

Skills: Go, Python, JavaScript, PostgreSQL, Docker, OpenCV, TensorFlow, REST APIs, Linux.
Services: AI Engineering, Backend Engineering, Full Stack Development, Technical Consulting.
Availability: Available for new projects.

Notable projects:
- Mall Surveillance System (Python, OpenCV, TensorFlow) — real-time object detection, face recognition, anomaly detection.
- AI Face Recognition System (Python, Deep Learning) — production-grade face recognition pipeline.
- Eye Disease Detection (Python, CNN, TensorFlow) — medical image classification.
- Hackerthon Platform (Go, PostgreSQL, JavaScript) — full-stack event platform.
- Text Analyzer (Python, NLP).
- CLI Calculator (Go).
- Gwinks Hub (Go, PostgreSQL, JavaScript) — full-stack platform.

Your job is to answer visitor questions about Victor's work, skills, experience, pricing, timelines,
and availability, and to encourage qualified visitors to start a project inquiry via the contact form
at /contact. Be concise (2-4 sentences), friendly, and specific — reference actual projects and
technologies where relevant. Do not invent details that aren't listed above. If asked something
unrelated to Victor's work (e.g. general trivia, coding help unrelated to his portfolio), gently steer
the conversation back to how Victor's services might help. Never claim to be human; you may say you're
Victor's AI assistant.`

type chatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type upstreamRequest struct {
	Model       string        `json:"model"`
	Messages    []chatMessage `json:"messages"`
	Temperature float64       `json:"temperature"`
	MaxTokens   int           `json:"max_tokens"`
}

type upstreamResponse struct {
	Choices []struct {
		Message chatMessage `json:"message"`
	} `json:"choices"`
	Error *struct {
		Message string `json:"message"`
	} `json:"error"`
}

// ChatRequest is the payload the frontend widget sends.
type ChatRequest struct {
	Message string        `json:"message"`
	History []chatMessage `json:"history"` // prior turns, oldest first, role: "user"|"assistant"
}

type ChatResponse struct {
	Reply string `json:"reply"`
}

var httpClient = &http.Client{Timeout: 20 * time.Second}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "...(truncated)"
}

// ChatHandler proxies portfolio-assistant chat turns to Ollama Cloud so the
// API key never reaches the browser.
func ChatHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req ChatRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	req.Message = strings.TrimSpace(req.Message)
	if req.Message == "" {
		http.Error(w, "Message is required", http.StatusBadRequest)
		return
	}
	if len(req.Message) > 2000 {
		http.Error(w, "Message too long", http.StatusBadRequest)
		return
	}

	apiKey := os.Getenv("OLLAMA_API_KEY")
	if apiKey == "" {
		http.Error(w, "Assistant is not configured", http.StatusServiceUnavailable)
		return
	}

	model := os.Getenv("OLLAMA_MODEL")
	if model == "" {
		model = defaultModel
	}

	messages := make([]chatMessage, 0, len(req.History)+2)
	messages = append(messages, chatMessage{Role: "system", Content: systemPrompt})

	// Cap history to the last 10 turns so the prompt (and bill) stays small.
	history := req.History
	if len(history) > 10 {
		history = history[len(history)-10:]
	}
	for _, m := range history {
		if m.Role != "user" && m.Role != "assistant" {
			continue
		}
		messages = append(messages, m)
	}
	messages = append(messages, chatMessage{Role: "user", Content: req.Message})

	reqBody, err := json.Marshal(upstreamRequest{
		Model:       model,
		Messages:    messages,
		Temperature: 0.6,
		MaxTokens:   300,
	})
	if err != nil {
		http.Error(w, "Failed to build request", http.StatusInternalServerError)
		return
	}

	httpReq, err := http.NewRequestWithContext(r.Context(), http.MethodPost, ollamaAPIURL, bytes.NewReader(reqBody))
	if err != nil {
		http.Error(w, "Failed to reach assistant", http.StatusInternalServerError)
		return
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+apiKey)

	resp, err := httpClient.Do(httpReq)
	if err != nil {
		http.Error(w, "Assistant is temporarily unavailable", http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		http.Error(w, "Failed to read assistant response", http.StatusBadGateway)
		return
	}

	log.Printf("[assistant] ollama status=%d body=%s", resp.StatusCode, truncate(string(body), 1000))

	if resp.StatusCode != http.StatusOK {
		http.Error(w, fmt.Sprintf("Assistant error: upstream returned status %d", resp.StatusCode), http.StatusBadGateway)
		return
	}

	var upstreamResp upstreamResponse
	if err := json.Unmarshal(body, &upstreamResp); err != nil {
		http.Error(w, "Failed to parse assistant response", http.StatusBadGateway)
		return
	}

	if len(upstreamResp.Choices) == 0 {
		http.Error(w, "Assistant returned no response", http.StatusBadGateway)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(ChatResponse{Reply: strings.TrimSpace(upstreamResp.Choices[0].Message.Content)})
}
