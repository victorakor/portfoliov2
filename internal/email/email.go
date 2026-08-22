package email

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
)

// Resend's transactional send endpoint.
const resendAPIURL = "https://api.resend.com/emails"

// defaultFrom must live on a domain that is verified for *sending* in Resend.
// victorakor.com is not verified; victorshackathon.xyz is (its DNS shows
// partially_failed, but capabilities.sending is enabled). Sending from an
// unverified domain fails with 403. Override with RESEND_FROM.
const defaultFrom = "Portfolio <portfolio@victorshackathon.xyz>"

// Where lead notifications land. Override with LEAD_NOTIFY_TO.
const defaultNotifyTo = "victorakor04@gmail.com"

var httpClient = &http.Client{Timeout: 15 * time.Second}

// Message is one outbound email. Text is required; HTML is optional.
type Message struct {
	To      string
	Subject string
	Text    string
	HTML    string
	ReplyTo string
}

type sendRequest struct {
	From    string   `json:"from"`
	To      []string `json:"to"`
	Subject string   `json:"subject"`
	Text    string   `json:"text,omitempty"`
	HTML    string   `json:"html,omitempty"`
	ReplyTo string   `json:"reply_to,omitempty"`
}

// Configured reports whether an API key is present. Callers can skip building a
// message when email is switched off.
func Configured() bool {
	return os.Getenv("RESEND_API_KEY") != ""
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "...(truncated)"
}

// Send delivers one message through Resend.
//
// With RESEND_API_KEY unset this is a no-op that returns nil, so local
// development and the self-hosted compose stack run fine without an email
// provider configured.
func Send(ctx context.Context, msg Message) error {
	apiKey := os.Getenv("RESEND_API_KEY")
	if apiKey == "" {
		return nil
	}

	from := os.Getenv("RESEND_FROM")
	if from == "" {
		from = defaultFrom
	}

	if strings.TrimSpace(msg.To) == "" {
		return fmt.Errorf("email: no recipient")
	}

	body, err := json.Marshal(sendRequest{
		From:    from,
		To:      []string{msg.To},
		Subject: msg.Subject,
		Text:    msg.Text,
		HTML:    msg.HTML,
		ReplyTo: msg.ReplyTo,
	})
	if err != nil {
		return fmt.Errorf("email: marshal: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, resendAPIURL, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("email: build request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey)

	resp, err := httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("email: send: %w", err)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)

	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		log.Printf("[email] resend status=%d body=%s", resp.StatusCode, truncate(string(respBody), 500))
		return fmt.Errorf("email: resend returned status %d", resp.StatusCode)
	}

	log.Printf("[email] sent subject=%q status=%d", msg.Subject, resp.StatusCode)
	return nil
}

// LeadNotification carries the contact-form fields worth putting in an email.
// It is declared here rather than reusing leads.Lead so that package leads can
// import package email without a cycle.
type LeadNotification struct {
	Name        string
	Email       string
	Phone       string
	Company     string
	ProjectType string
	Budget      string
	Timeline    string
	Message     string
}

func orDash(s string) string {
	if strings.TrimSpace(s) == "" {
		return "—"
	}
	return s
}

// NotifyNewLead emails the site owner about a fresh contact-form submission.
// ReplyTo is set to the lead's address, so replying from the inbox goes straight
// to them instead of to the sending domain.
func NotifyNewLead(ctx context.Context, lead LeadNotification) error {
	to := os.Getenv("LEAD_NOTIFY_TO")
	if to == "" {
		to = defaultNotifyTo
	}

	subject := fmt.Sprintf("New lead: %s", lead.Name)
	if strings.TrimSpace(lead.Company) != "" {
		subject = fmt.Sprintf("New lead: %s (%s)", lead.Name, lead.Company)
	}

	fields := [][2]string{
		{"Name", orDash(lead.Name)},
		{"Email", orDash(lead.Email)},
		{"Phone", orDash(lead.Phone)},
		{"Company", orDash(lead.Company)},
		{"Project type", orDash(lead.ProjectType)},
		{"Budget", orDash(lead.Budget)},
		{"Timeline", orDash(lead.Timeline)},
	}

	var text strings.Builder
	for _, f := range fields {
		fmt.Fprintf(&text, "%-13s %s\n", f[0]+":", f[1])
	}
	fmt.Fprintf(&text, "\nMessage:\n%s\n", orDash(lead.Message))
	fmt.Fprintf(&text, "\nManage: https://victorakor.onrender.com/admin/leads\n")

	// Lead content is attacker-controlled, so escape it rather than interpolating
	// raw strings into the HTML body.
	var html strings.Builder
	html.WriteString(`<div style="font-family:system-ui,-apple-system,Segoe UI,sans-serif;font-size:14px;line-height:1.6">`)
	html.WriteString(`<h2 style="margin:0 0 16px">New project inquiry</h2><table cellpadding="4" style="border-collapse:collapse">`)
	for _, f := range fields {
		fmt.Fprintf(&html, `<tr><td style="color:#666">%s</td><td><strong>%s</strong></td></tr>`,
			template.HTMLEscapeString(f[0]), template.HTMLEscapeString(f[1]))
	}
	html.WriteString(`</table>`)
	fmt.Fprintf(&html, `<p style="color:#666;margin:16px 0 4px">Message</p><div style="white-space:pre-wrap;padding:12px;background:#f5f5f5;border-radius:6px">%s</div>`,
		template.HTMLEscapeString(orDash(lead.Message)))
	html.WriteString(`<p style="margin-top:20px"><a href="https://victorakor.onrender.com/admin/leads">Open the leads dashboard</a></p></div>`)

	return Send(ctx, Message{
		To:      to,
		Subject: subject,
		Text:    text.String(),
		HTML:    html.String(),
		ReplyTo: lead.Email,
	})
}
