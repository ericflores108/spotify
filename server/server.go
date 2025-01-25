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

	return mux
}
