package handlers

import (
	"github.com/gorilla/mux"
	"io"
	"net/http"
	"os"
	"transmuxer/pkg/host"
	"transmuxer/pkg/utils"
)

func PlaylistHandler(w http.ResponseWriter, r *http.Request) {
	var err error
	vars := mux.Vars(r)
	filename := vars["name"]
	q := r.URL.Query()
	u := q.Get("u")
	cu := q.Get("cu")
	if !utils.IsNameAdmissible(filename) {
		http.Error(w, "invalid path variable 'name'", http.StatusBadRequest)
		return
	}

	if u == "" {
		// FIXME: Dangerous, add a secret key and hash filename with it instead
		if _, err = os.Stat(host.GetPlaylistPath(filename)); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		if _, err = os.Stat(host.GetSegmentsDir(filename)); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
	}
	p := host.GetPlaylistPath(filename)
	if _, err = os.Stat(p); err != nil {
		if err = host.GeneratePlaylist(host.PlayInput{
			PlaylistURL: u,
			CaptionsURL: cu,
			Filename:    filename,
		}); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	w.Header().Set("Content-Type", "application/vnd.apple.mpegurl\"")
	file, err := os.Open(p)
	defer file.Close()
	if err != nil {
		w.Write([]byte("no file exists matching input name"))
		return
	}
	_, err = io.Copy(w, file)
	if err != nil {
		http.Error(w, "Error writing file contents to response", http.StatusInternalServerError)
		return
	}
}
