package main

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"strings"
)

func generateID() (string, error) {
	b := make([]byte, 6) // генерирует 8 символов после base64 кодирования
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(b), nil
}

func generateShortUrl(storage URLStorage, res http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodPost {
		http.Error(res, "Only POST requests are allowed!", http.StatusBadRequest)
		return
	}
	url, err := io.ReadAll(req.Body)
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {

		}
	}(req.Body) // Важно закрыть тело запроса

	if err != nil {
		http.Error(res, err.Error(), http.StatusBadRequest)
		return
	}
	id, _ := generateID()
	storage.AddURL(id, string(url))

	res.WriteHeader(http.StatusCreated)
	_, err = fmt.Fprintf(res, "http://localhost:8080/%s", id)
	if err != nil {
		return
	}
}

func getShortUrl(storage URLStorage, res http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodGet {
		http.Error(res, "Only GET requests are allowed!", http.StatusBadRequest)
		return
	}
	id := strings.TrimPrefix(req.URL.Path, "/")
	url, ok := storage.GetURL(id)
	if !ok {
		http.Error(res, "", http.StatusBadRequest)
		return
	}
	http.Redirect(res, req, url, http.StatusTemporaryRedirect)
}

func handleRequests(storage URLStorage, res http.ResponseWriter, req *http.Request) {
	url := req.URL.Path
	if url == "/" {
		generateShortUrl(storage, res, req)
		return
	}
	getShortUrl(storage, res, req)
}

func main() {
	storage := NewMemoryURLStorage()
	mux := http.NewServeMux()
	mux.HandleFunc(`/`, func(w http.ResponseWriter, r *http.Request) {
		handleRequests(storage, w, r)
	})

	err := http.ListenAndServe(`:8080`, mux)
	if err != nil {
		panic(err)
	}
}
