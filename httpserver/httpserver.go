package httpserver

import (
	"context"
	"html/template"
	"net/http"

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

	mux.HandleFunc("/addToPlaylist", func(w http.ResponseWriter, r *http.Request) {
		s.Service.AddToPlaylistHandler(w, s.Ctx)
	})

	mux.HandleFunc("/generatePlaylist", func(w http.ResponseWriter, r *http.Request) {
		queryParams := r.URL.Query()
		albumID := queryParams.Get("albumID")
		userID := queryParams.Get("userID")

		if albumID == "" || userID == "" {
			http.Error(w, "Missing albumName or userID query parameters", http.StatusBadRequest)
			return
		}

		s.Service.GeneratePlaylistHandler(w, s.Ctx, albumID, userID)
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
