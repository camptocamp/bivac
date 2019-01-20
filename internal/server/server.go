package server

/*
import (
	"fmt"
	"net/http"

	log "github.com/Sirupsen/logrus"
	"github.com/gorilla/mux"

	"github.com/camptocamp/bivac/internal/manager"
	"github.com/camptocamp/bivac/pkg/orchestrators"
)

type Server struct {
	Address string
	PSK     string
}

func Start(m *manager.Manager) (err error) {
	router := mux.NewRouter().StrictSlash(true)

	router.Handle("/volumes", m.Server.handleAPIRequest(http.HandlerFunc(m.Server.getVolumes)))
	router.Handle("/ping", m.Server.handleAPIRequest(http.HandlerFunc(m.Server.ping)))

	log.Infof("Listening on %s", m.Server.Address)
	log.Fatal(http.ListenAndServe(m.Server.Address, router))
	return
}

func (s *Server) handleAPIRequest(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != fmt.Sprintf("Bearer %s", s.PSK) {
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte("Unauthorized"))
			return
		}
		next.ServeHTTP(w, r)
	})
}

func (s *Server) getVolumes(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
}

func (s *Server) ping(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"type":"pong"}`))
	return
}
*/
