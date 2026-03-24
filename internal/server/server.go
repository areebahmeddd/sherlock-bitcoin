package server

import "net/http"

func New() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/health", handleHealth)
	mux.HandleFunc("/api/blocks", handleListBlocks)
	mux.HandleFunc("/api/blocks/", handleBlocksDispatch)

	fileServer := http.FileServer(http.Dir("docs"))
	mux.Handle("/docs/", http.StripPrefix("/docs/", fileServer))
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		http.ServeFile(w, r, "docs/index.html")
	})

	return mux
}
