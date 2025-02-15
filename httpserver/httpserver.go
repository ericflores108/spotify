package httpserver

import (
	"context"
	"fmt"
	"html/template"
	"net/http"
	"strings"

	"github.com/ericflores108/spotify/htmlpages"
	"github.com/ericflores108/spotify/logger"
	"github.com/ericflores108/spotify/service"
)

type Server struct {
	Service *service.Service
}

func NewServer(service *service.Service) *Server {
	return &Server{
		Service: service,
	}
}

func (s *Server) RegisterRoutes(ctx context.Context) *http.ServeMux {
	mux := http.NewServeMux()

	// Serve static files
	mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("./static"))))

	tmpl := template.Must(template.New("index").Parse(htmlpages.Home))

	// Serve the root page with the template
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		err := tmpl.Execute(w, nil)
		if err != nil {
			http.Error(w, "Failed to render template", http.StatusInternalServerError)
		}
	})

	// Authentication routes
	mux.HandleFunc("/login", s.Service.LoginHandler)
	mux.HandleFunc("/callback", func(w http.ResponseWriter, r *http.Request) {
		s.Service.CallbackHandler(w, ctx, r)
	})

	// Protected routes with middleware
	mux.Handle("/home", service.CookieConsentMiddleware(http.HandlerFunc(s.Service.HomePageHandler)))
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

		s.Service.GeneratePlaylistHandler(w, ctx, parts[1], userID, accessToken, r)
	})

	return mux
}
