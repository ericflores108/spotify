package server

import (
	"context"
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

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Welcome to the Spotify Service!"))
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

	mux.HandleFunc("/getAlbumDetails", func(w http.ResponseWriter, r *http.Request) {
		s.Service.GetAlbumDetailsHandler(w, s.Ctx)
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

	return mux
}
