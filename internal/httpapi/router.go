package httpapi

import (
	"encoding/json"
	"log"
	"math/rand"
	"net/http"
	"strings"
	"url-shortener/internal/storage"
	"time"
)

type Router struct {
	db *storage.Postgres
}

func NewRouter(db *storage.Postgres) *http.ServeMux {
	r := &Router{db: db}
	mux := http.NewServeMux()

	mux.HandleFunc("/shorten", r.handleShorten)         // POST /shorten
	mux.HandleFunc("/stats/", r.handleStats)           // GET /stats/{code}
	mux.HandleFunc("/", r.handleRedirect)              // GET /{code}

	return mux
}

// ---- Response structs ----

type ShortenResponse struct {
	ShortURL string `json:"short_url"`
}

type StatsResponse struct {
	ShortURL string `json:"short_url"`
	Clicks   int    `json:"clicks"`
}

// ---- Handlers ----

func (r *Router) handleShorten(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	type request struct {
		URL string `json:"url"`
	}
	var body request
	if err := json.NewDecoder(req.Body).Decode(&body); err != nil || body.URL == "" {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}

	short := r.generateCode()
	if err := r.db.Save(short, body.URL); err != nil {
		log.Println("Failed to save URL:", err)
		http.Error(w, "failed to save", http.StatusInternalServerError)
		return
	}

	resp := ShortenResponse{ShortURL: "http://localhost:8080/" + short}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (r *Router) handleRedirect(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	code := strings.TrimPrefix(req.URL.Path, "/")
	if code == "" || code == "shorten" || strings.HasPrefix(code, "stats") {
		http.NotFound(w, req)
		return
	}

	original, err := r.db.Find(code)
	if err != nil {
		http.NotFound(w, req)
		return
	}

	if err := r.db.IncrementClicks(code); err != nil {
		log.Println("Failed to increment clicks:", err)
	}

	log.Printf("Redirect: %s -> %s", code, original)
	http.Redirect(w, req, original, http.StatusFound)
}

func (r *Router) handleStats(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	code := strings.TrimPrefix(req.URL.Path, "/stats/")
	if code == "" {
		http.NotFound(w, req)
		return
	}

	clicks, err := r.db.GetClicks(code)
	if err != nil {
		http.NotFound(w, req)
		return
	}

	resp := StatsResponse{ShortURL: code, Clicks: clicks}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// ---- Helpers ----

func (r *Router) generateCode() string {
	const letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, 6)
	rand.Seed(time.Now().UnixNano())
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}
