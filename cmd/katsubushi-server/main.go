package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"net"
	"net/http"
	"strconv"
	"strings"
	"time"

	katsubushi "github.com/shogo82148/go-katsubushi-katshubushi"
	"gopkg.in/Sirupsen/logrus.v0"
)

type Server struct {
	s katsubushi.IdGenerator
}

func NewServer(s katsubushi.IdGenerator) *Server {
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
		logrus.WithFields(logrus.Fields{
			"id":          info.Id,
			"released_at": info.ReleasedAt,
			"expire_at":   info.ExpireAt,
		}).Info("put")
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
		logrus.WithFields(logrus.Fields{
			"id":          info.Id,
			"released_at": info.ReleasedAt,
			"expire_at":   info.ExpireAt,
		}).Info("put")
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
		logrus.WithFields(logrus.Fields{
			"id": id,
		}).Info("delete")
	}
}

func (s *Server) renderError(err error, w http.ResponseWriter) {
	encoder := json.NewEncoder(w)
	encoder.Encode(&struct {
		Error string `json:"error"`
	}{
		Error: err.Error(),
	})
	logrus.Error(err)
}

func main() {
	var host string
	var port int
	var expire time.Duration
	flag.StringVar(&host, "host", "", "bind hostname")
	flag.IntVar(&port, "port", 8080, "bind port")
	flag.DurationVar(&expire, "expire", 24*time.Hour, "expire time")
	flag.Parse()

	server := katsubushi.NewMemoryServer()
	server.ExpireDuration = expire
	http.Handle("/", NewServer(server))

	logrus.WithFields(logrus.Fields{
		"host":   host,
		"port":   port,
		"expire": expire,
	}).Info("start")

	http.ListenAndServe(net.JoinHostPort(host, strconv.Itoa(port)), nil)
}
