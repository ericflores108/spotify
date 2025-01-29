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
	Ctx     context.Context
}

func NewServer(ctx context.Context, service *service.Service) *Server {
	return &Server{
		Service: service,
		Ctx:     ctx,
	}
}

func (s *Server) RegisterRoutes() *http.ServeMux {
	mux := http.NewServeMux()

	// Serve static files (e.g., favicon.ico) from the ./static directory
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

	mux.HandleFunc("/login", s.Service.LoginHandler)

	mux.HandleFunc("/callback", func(w http.ResponseWriter, r *http.Request) {
		s.Service.CallbackHandler(w, s.Ctx, r)
	})

	mux.HandleFunc("/generatePlaylist", func(w http.ResponseWriter, r *http.Request) {
		// Ensure the request method is POST
		if r.Method != http.MethodPost {
			http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
			return
		}

		// Parse the form data
		if err := r.ParseForm(); err != nil {
			http.Error(w, "Failed to parse form", http.StatusBadRequest)
			return
		}

		// Extract the userID and album name from the form
		userID := r.FormValue("userID")
		albumURL := r.FormValue("albumURL")
		accessToken := r.FormValue("accessToken")

		if userID == "" {
			http.Error(w, "User ID cannot be empty", http.StatusBadRequest)
			return
		}

		// Set the logger prefix to the user ID
		logger.InfoLogger.SetPrefix(fmt.Sprintf("UserID: %s", userID))
		logger.DebugLogger.SetPrefix(fmt.Sprintf("UserID: %s", userID))
		logger.ErrorLogger.SetPrefix(fmt.Sprintf("UserID: %s", userID))

		if albumURL == "" {
			http.Error(w, "Album link cannot be empty", http.StatusBadRequest)
			return
		}

		if accessToken == "" {
			http.Error(w, "Access Token cannot be empty", http.StatusBadRequest)
			return
		}

		// Process the form submission
		logger.LogInfo("User ID submitted: %s", userID)
		logger.LogInfo("Album link submitted: %s", albumURL)

		parts := strings.Split(albumURL, "/album/")
		if len(parts) < 2 {
			http.Error(w, "Invalid URL format", http.StatusBadRequest)
			return
		}
		logger.LogDebug("Album link submitted: %s", albumURL)

		s.Service.GeneratePlaylistHandler(w, s.Ctx, parts[1], userID, accessToken, r)
	})

	return mux
}
