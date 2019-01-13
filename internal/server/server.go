package server

import (
	"fmt"
	"net/http"

	log "github.com/Sirupsen/logrus"
	"github.com/gorilla/mux"

	//"github.com/camptocamp/bivac/cmd/manager"
	"github.com/camptocamp/bivac/pkg/orchestrators"
)

type Server struct {
	Address string
	PSK     string
	o       orchestrators.Orchestrator
}

func Start(o orchestrators.Orchestrator, s *Server) (err error) {
	s.o = o

	router := mux.NewRouter().StrictSlash(true)

	router.Handle("/volumes", s.handleAPIRequest(http.HandlerFunc(s.getVolumes)))

	log.Infof("Listening on %s", s.Address)
	log.Fatal(http.ListenAndServe(s.Address, router))
	return
}

func (s *Server) handleAPIRequest(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != fmt.Sprintf("Bearer %s", s.PSK) {
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte("403 - Unauthorized"))
			return
		}
		next.ServeHTTP(w, r)
	})
}

func (s *Server) getVolumes(w http.ResponseWriter, r *http.Request) {
	o.GetVolumes()
	w.WriteHeader(http.StatusOK)
}
