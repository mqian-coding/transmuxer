package handlers

import (
	"concurrency-practice/pkg/utils"
	"github.com/gorilla/mux"
	"io"
	"net/http"
	"os"
)

func PlaylistHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	filename := vars["name"]
	if !utils.IsNameAdmissible(filename) {
		w.Write([]byte("invalid path variable 'name'"))
		return
	}
	w.Header().Set("Content-Type", "application/vnd.apple.mpegurl\"")
	file, err := os.Open("server/playlists/" + filename)
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
