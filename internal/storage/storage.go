package storage

import (
	"bytes"
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"path"
	"sort"
	"strconv"
	"strings"
	"time"
)

// Uploads land in this Cloudinary folder.
const uploadFolder = "portfolio"

// Render's free tier has an ephemeral filesystem, so files are streamed straight
// to Cloudinary and never written to disk.
const maxUploadBytes = 10 << 20 // 10 MB

// Sniffed content types we accept. The browser-supplied Content-Type and the
// filename extension are both attacker-controlled, so neither is trusted.
var allowedTypes = map[string]bool{
	"image/jpeg": true,
	"image/png":  true,
	"image/webp": true,
	"image/gif":  true,
}

var httpClient = &http.Client{Timeout: 60 * time.Second}

type config struct {
	cloudName string
	apiKey    string
	apiSecret string
}

// loadConfig reads CLOUDINARY_URL (cloudinary://api_key:api_secret@cloud_name),
// falling back to discrete vars.
func loadConfig() (config, error) {
	if raw := strings.TrimSpace(os.Getenv("CLOUDINARY_URL")); raw != "" {
		u, err := url.Parse(raw)
		if err != nil {
			return config{}, fmt.Errorf("storage: CLOUDINARY_URL is not a valid URL: %w", err)
		}
		secret, _ := u.User.Password()
		c := config{
			cloudName: u.Host,
			apiKey:    u.User.Username(),
			apiSecret: secret,
		}
		if c.cloudName == "" || c.apiKey == "" || c.apiSecret == "" {
			return config{}, fmt.Errorf("storage: CLOUDINARY_URL is missing the cloud name, key, or secret")
		}
		return c, nil
	}

	c := config{
		cloudName: os.Getenv("CLOUDINARY_CLOUD_NAME"),
		apiKey:    os.Getenv("CLOUDINARY_API_KEY"),
		apiSecret: os.Getenv("CLOUDINARY_SECRET"),
	}
	if c.cloudName == "" || c.apiKey == "" || c.apiSecret == "" {
		return config{}, fmt.Errorf("storage: Cloudinary is not configured")
	}
	return c, nil
}

// sign implements Cloudinary's signed-upload scheme: the signed parameters
// sorted by key as k=v&k=v, the API secret appended with no separator, SHA-1
// hex. Signing here rather than using an unsigned browser preset keeps the
// secret server-side.
func sign(params map[string]string, apiSecret string) string {
	keys := make([]string, 0, len(params))
	for k := range params {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	parts := make([]string, 0, len(keys))
	for _, k := range keys {
		parts = append(parts, k+"="+params[k])
	}

	sum := sha1.Sum([]byte(strings.Join(parts, "&") + apiSecret))
	return hex.EncodeToString(sum[:])
}

type uploadResponse struct {
	SecureURL string `json:"secure_url"`
	PublicID  string `json:"public_id"`
	Width     int    `json:"width"`
	Height    int    `json:"height"`
	Format    string `json:"format"`
	Bytes     int    `json:"bytes"`
	Error     *struct {
		Message string `json:"message"`
	} `json:"error"`
}

// UploadHandler accepts a multipart "file" field and returns the resulting
// Cloudinary URL. It is mounted under /api/admin/, so middleware.Auth already
// requires a valid admin token.
func UploadHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	cfg, err := loadConfig()
	if err != nil {
		log.Printf("[storage] %v", err)
		http.Error(w, "Image uploads are not configured", http.StatusServiceUnavailable)
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, maxUploadBytes)
	if err := r.ParseMultipartForm(maxUploadBytes); err != nil {
		http.Error(w, "File is too large (max 10 MB)", http.StatusRequestEntityTooLarge)
		return
	}
	defer r.MultipartForm.RemoveAll()

	file, header, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "No file provided", http.StatusBadRequest)
		return
	}
	defer file.Close()

	// Sniff the real type from the leading bytes, then rewind so the whole file
	// still gets uploaded.
	sniff := make([]byte, 512)
	n, _ := io.ReadFull(file, sniff)
	if n == 0 {
		http.Error(w, "File is empty", http.StatusBadRequest)
		return
	}
	detected := strings.SplitN(http.DetectContentType(sniff[:n]), ";", 2)[0]
	if !allowedTypes[detected] {
		http.Error(w, fmt.Sprintf("Unsupported file type %q — use JPEG, PNG, WebP, or GIF", detected), http.StatusBadRequest)
		return
	}
	if _, err := file.Seek(0, io.SeekStart); err != nil {
		http.Error(w, "Could not read the file", http.StatusInternalServerError)
		return
	}

	timestamp := strconv.FormatInt(time.Now().Unix(), 10)
	signed := map[string]string{
		"folder":    uploadFolder,
		"timestamp": timestamp,
	}

	var body bytes.Buffer
	mw := multipart.NewWriter(&body)
	for k, v := range signed {
		if err := mw.WriteField(k, v); err != nil {
			http.Error(w, "Could not build the upload", http.StatusInternalServerError)
			return
		}
	}
	// api_key and signature are sent but never themselves signed.
	if err := mw.WriteField("api_key", cfg.apiKey); err != nil {
		http.Error(w, "Could not build the upload", http.StatusInternalServerError)
		return
	}
	if err := mw.WriteField("signature", sign(signed, cfg.apiSecret)); err != nil {
		http.Error(w, "Could not build the upload", http.StatusInternalServerError)
		return
	}

	part, err := mw.CreateFormFile("file", path.Base(header.Filename))
	if err != nil {
		http.Error(w, "Could not build the upload", http.StatusInternalServerError)
		return
	}
	if _, err := io.Copy(part, file); err != nil {
		http.Error(w, "Could not read the file", http.StatusInternalServerError)
		return
	}
	if err := mw.Close(); err != nil {
		http.Error(w, "Could not build the upload", http.StatusInternalServerError)
		return
	}

	endpoint := "https://api.cloudinary.com/v1_1/" + cfg.cloudName + "/image/upload"
	req, err := http.NewRequestWithContext(r.Context(), http.MethodPost, endpoint, &body)
	if err != nil {
		http.Error(w, "Could not reach the image host", http.StatusInternalServerError)
		return
	}
	req.Header.Set("Content-Type", mw.FormDataContentType())

	resp, err := httpClient.Do(req)
	if err != nil {
		log.Printf("[storage] cloudinary request failed: %v", err)
		http.Error(w, "Image host is temporarily unavailable", http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		http.Error(w, "Could not read the image host response", http.StatusBadGateway)
		return
	}

	var parsed uploadResponse
	if err := json.Unmarshal(respBody, &parsed); err != nil {
		log.Printf("[storage] cloudinary status=%d unparseable body", resp.StatusCode)
		http.Error(w, "Could not parse the image host response", http.StatusBadGateway)
		return
	}

	if resp.StatusCode != http.StatusOK || parsed.SecureURL == "" {
		detail := "upload rejected"
		if parsed.Error != nil && parsed.Error.Message != "" {
			detail = parsed.Error.Message
		}
		log.Printf("[storage] cloudinary status=%d error=%s", resp.StatusCode, detail)
		http.Error(w, "Upload failed: "+detail, http.StatusBadGateway)
		return
	}

	log.Printf("[storage] uploaded public_id=%s bytes=%d type=%s", parsed.PublicID, parsed.Bytes, detected)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{
		"url":       parsed.SecureURL,
		"public_id": parsed.PublicID,
		"width":     parsed.Width,
		"height":    parsed.Height,
		"format":    parsed.Format,
		"bytes":     parsed.Bytes,
	})
}
