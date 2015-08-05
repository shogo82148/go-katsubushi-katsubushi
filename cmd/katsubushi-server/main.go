package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"

	katsubushi "github.com/shogo82148/go-katsubushi-katshubushi"
)

type Server struct {
	s katsubushi.Server
}

func NewServer(s katsubushi.Server) *Server {
	return &Server{s}
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	switch r.Method {
	case "POST":
		// create new id
		info, err := s.s.New()
		if err != nil {
			s.renderError(err, w)
			return
		}
		encoder := json.NewEncoder(w)
		encoder.Encode(info)
	case "PUT":
		// extend expire time of id
		path := strings.Split(r.URL.Path, "/")
		id, err := strconv.Atoi(path[len(path)-1])
		if err != nil {
			s.renderError(err, w)
			return
		}

		info, err := s.s.Update(id)
		if err != nil {
			s.renderError(err, w)
			return
		}

		encoder := json.NewEncoder(w)
		encoder.Encode(info)
	case "DELETE":
		// delete id
		path := strings.Split(r.URL.Path, "/")
		id, err := strconv.Atoi(path[len(path)-1])
		if err != nil {
			s.renderError(err, w)
			return
		}
		s.s.Delete(id)
		fmt.Fprintf(w, `{"deleted":true}`)
	}
}

func (s *Server) renderError(err error, w http.ResponseWriter) {
	encoder := json.NewEncoder(w)
	encoder.Encode(&struct {
		Error string `json:"error"`
	}{
		Error: err.Error(),
	})
	log.Printf("error: %s", err)
}

func main() {
	flag.Parse()

	http.Handle("/", NewServer(katsubushi.NewMemoryServer()))
	http.ListenAndServe(":8080", nil)
}
