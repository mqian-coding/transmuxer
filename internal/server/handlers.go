package server

import (
	"concurrency-practice/internal/server/handlers"
	"errors"
	"github.com/gorilla/mux"
)

func registerHandlers(r *mux.Router) error {
	if r == nil {
		return errors.New("router is not initialized")
	}
	r.HandleFunc("/{name}/playlist.m3u8", handlers.PlaylistHandler)
	r.HandleFunc("/transmux", handlers.TransmuxHandler)
	return nil
}
