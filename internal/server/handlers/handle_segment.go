package handlers

import (
	"concurrency-practice/pkg/host"
	"concurrency-practice/pkg/utils"
	"github.com/gorilla/mux"
	"io"
	"net/http"
	"os"
)

func SegmentHandler(w http.ResponseWriter, r *http.Request) {
	var err error
	vars := mux.Vars(r)
	name := vars["name"]
	segName := vars["seg_name"]
	if !utils.IsNameAdmissible(name) {
		http.Error(w, "invalid path variable 'name'", http.StatusBadRequest)
		return
	}
	if !utils.IsNameAdmissible(segName) {
		http.Error(w, "invalid path variable 'seg_name'", http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/octet-stream\"")
	file, err := os.Open(host.GetSegmentsPath(name, segName))
	defer file.Close()
	if err != nil {
		w.Write([]byte("no segment exists matching input name"))
		return
	}
	_, err = io.Copy(w, file)
	if err != nil {
		http.Error(w, "Error writing segment contents to response", http.StatusInternalServerError)
		return
	}
}
