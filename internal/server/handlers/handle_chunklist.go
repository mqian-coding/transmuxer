package handlers

import (
	"concurrency-practice/pkg/host"
	"concurrency-practice/pkg/utils"
	"github.com/gorilla/mux"
	"io"
	"net/http"
	"os"
)

func ChunklistHandler(w http.ResponseWriter, r *http.Request) {
	var err error
	vars := mux.Vars(r)
	name := vars["name"]
	if !utils.IsNameAdmissible(name) {
		http.Error(w, "invalid path variable 'name'", http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/vnd.apple.mpegurl\"")
	file, err := os.Open(host.GetChunklistPath(name))
	defer file.Close()
	if err != nil {
		w.Write([]byte("no chunklist manifest exists matching input name"))
		return
	}
	_, err = io.Copy(w, file)
	if err != nil {
		http.Error(w, "Error writing manifest contents to response", http.StatusInternalServerError)
		return
	}
}
