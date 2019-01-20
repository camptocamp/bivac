package manager

import (
	"encoding/json"
	"fmt"
	"net/http"

	log "github.com/Sirupsen/logrus"
	"github.com/gorilla/mux"
	//"github.com/camptocamp/bivac/pkg/orchestrators"
)

type Server struct {
	Address string
	PSK     string
}

func (m *Manager) StartServer() (err error) {
	router := mux.NewRouter().StrictSlash(true)

	router.Handle("/volumes", m.handleAPIRequest(http.HandlerFunc(m.getVolumes)))
	router.Handle("/ping", m.handleAPIRequest(http.HandlerFunc(m.ping)))

	log.Infof("Listening on %s", m.Server.Address)
	log.Fatal(http.ListenAndServe(m.Server.Address, router))
	return
}

func (m *Manager) handleAPIRequest(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != fmt.Sprintf("Bearer %s", m.Server.PSK) {
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte("Unauthorized"))
			return
		}
		next.ServeHTTP(w, r)
	})
}

func (m *Manager) getVolumes(w http.ResponseWriter, r *http.Request) {
	b, err := json.Marshal(m.Volumes)
	if err != nil {
		log.Errorf("failed to marshal volumes: %s", err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("500 - Internal server error"))
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write(b)
	return
}

func (m *Manager) ping(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"type":"pong"}`))
	return
}
