package handlers

import (
	"concurrency-practice/pkg/transmuxer"
	"concurrency-practice/pkg/utils"
	"fmt"
	"log"
	"net/http"
	"net/url"
)

func TransmuxHandler(w http.ResponseWriter, r *http.Request) {
	var n, u string
	qp := r.URL.Query()
	if u = qp.Get("u"); u == "" {
		http.Error(w, "query param 'u' must be url pointing to master manifest", http.StatusBadRequest)
		return
	}
	u, err := url.QueryUnescape(u)
	if err != nil {
		http.Error(w, "query param 'u' must be url with valid url safe encoding", http.StatusBadRequest)
		return
	}
	if n = qp.Get("n"); n == "" {
		http.Error(w, "query param 'n' is needed to name output file", http.StatusBadRequest)
		return
	}
	if ok := utils.IsNameAdmissible(n); !ok {
		http.Error(w, "query param 'n' must be alphanumeric, underscore delimited", http.StatusBadRequest)
		return
	}

	if err = transmuxer.Transmux(transmuxer.TransmuxInput{
		PlaylistURL: u,
		OutputName:  n,
	}); err != nil {
		log.Printf(fmt.Sprintf("transmux request failed: %s", err.Error()))
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
}
