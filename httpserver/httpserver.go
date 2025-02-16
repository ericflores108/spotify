package httpserver

import (
	"context"
	"fmt"
	"html/template"
	"net/http"
	"strings"

	"github.com/ericflores108/spotify/handlers"
	"github.com/ericflores108/spotify/htmlpages"
	"github.com/ericflores108/spotify/logger"
)

type Server struct {
	Handler *handlers.Service
}

func NewServer(handler *handlers.Service) *Server {
	return &Server{
		Handler: handler,
	}
}

func (s *Server) RegisterRoutes(ctx context.Context) *http.ServeMux {
	mux := http.NewServeMux()

	// Serve static files
	mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("./static"))))

	tmpl := template.Must(template.New("index").Parse(htmlpages.Login))

	// Serve the root page with the template
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		err := tmpl.Execute(w, nil)
		if err != nil {
			http.Error(w, "Failed to render template", http.StatusInternalServerError)
		}
	})

	// Authentication routes
	mux.HandleFunc("/login", s.Handler.LoginHandler)
	mux.HandleFunc("/callback", func(w http.ResponseWriter, r *http.Request) {
		s.Handler.CallbackHandler(w, ctx, r)
	})

	// Spotify route doesn't need auth. It creates a playlist only in eflorty
	mux.HandleFunc("/spotify", func(w http.ResponseWriter, r *http.Request) {
		s.Handler.HomePageHandler(w, ctx, r)
	})

	// Protected routes with middleware
	mux.HandleFunc("/home", func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie("cookies_accepted")
		if err != nil || cookie.Value == "false" {
			http.Error(w, "You must accept cookies to use this site.", http.StatusForbidden)
			return
		}
		s.Handler.HomePageHandler(w, ctx, r)
	})
	mux.HandleFunc("/generatePlaylist", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
			return
		}

		if err := r.ParseForm(); err != nil {
			http.Error(w, "Failed to parse form", http.StatusBadRequest)
			return
		}

		userID := r.FormValue("userID")
		albumURL := r.FormValue("albumURL")
		accessToken := r.FormValue("accessToken")

		if userID == "" || albumURL == "" || accessToken == "" {
			http.Error(w, "Missing required fields", http.StatusBadRequest)
			return
		}

		logger.InfoLogger.SetPrefix(fmt.Sprintf("UserID: %s", userID))
		logger.DebugLogger.SetPrefix(fmt.Sprintf("UserID: %s", userID))
		logger.ErrorLogger.SetPrefix(fmt.Sprintf("UserID: %s", userID))

		logger.LogInfo("Album link submitted: %s", albumURL)

		parts := strings.Split(albumURL, "/album/")
		if len(parts) < 2 {
			http.Error(w, "Invalid URL format", http.StatusBadRequest)
			return
		}

		s.Handler.GeneratePlaylistHandler(w, ctx, parts[1], userID, accessToken, r)
	})

	return mux
}
