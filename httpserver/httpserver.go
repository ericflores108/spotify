package httpserver

import (
	"context"
	"html/template"
	"net/http"
	"strings"

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

	// Define the template for the root page
	tmpl := template.Must(template.New("index").Parse(`
<!doctype html>
<html>
<head>
	<title>Authorization Code Flow Example</title>
</head>
<body>
	<div>
		<a href="/login">Log in with Spotify</a>
	</div>
</body>
</html>`))

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

	mux.HandleFunc("/topTracks", func(w http.ResponseWriter, r *http.Request) {
		s.Service.StoreTracksHandler(w, s.Ctx)
	})

	mux.HandleFunc("/createPlaylist", func(w http.ResponseWriter, r *http.Request) {
		s.Service.CreatePlaylistHandler(w, s.Ctx)
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

		if userID == "" {
			http.Error(w, "User ID cannot be empty", http.StatusBadRequest)
			return
		}
		if albumURL == "" {
			http.Error(w, "Album link cannot be empty", http.StatusBadRequest)
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

		s.Service.GeneratePlaylistHandler(w, s.Ctx, parts[1], userID, r)
	})

	mux.HandleFunc("/createUser", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
			return
		}
		s.Service.CreateUserHandler(w, r)
	})

	return mux
}
