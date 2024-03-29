package server

import (
	"errors"
	"github.com/gorilla/mux"
	"transmuxer/internal/server/handlers"
)

func registerHandlers(r *mux.Router) error {
	if r == nil {
		return errors.New("router is not initialized")
	}
	r.HandleFunc("/{name}/playlist.m3u8", handlers.PlaylistHandler)
	r.HandleFunc("/{name}/chunklist.m3u8", handlers.ChunklistHandler)
	r.HandleFunc("/{name}/{seg_name}/seg.ts", handlers.SegmentHandler)
	r.HandleFunc("/transmux", handlers.TransmuxHandler)
	return nil
}
