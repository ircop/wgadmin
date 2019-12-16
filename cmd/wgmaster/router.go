package main

import (
	"bytes"
	"encoding/base64"
	"net/http"
	"strings"

	"github.com/julienschmidt/httprouter"
)

func basicAuth(h httprouter.Handle, login, password []byte) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		if len(login) == 0 || len(password) == 0 {
			h(w, r, ps)
			return
		}

		const basicAuthPrefix string = "Basic "

		auth := r.Header.Get("Authorization")
		if strings.HasPrefix(auth, basicAuthPrefix) {
			payload, err := base64.StdEncoding.DecodeString(auth[len(basicAuthPrefix):])
			if err == nil {
				pair := bytes.SplitN(payload, []byte(":"), 2)
				if len(pair) == 2 && bytes.Equal(pair[0], login) && bytes.Equal(pair[1], password) {
					// Delegate request to the given handle
					h(w, r, ps)
					return
				}
			}
		}

		w.Header().Set("WWW-Authenticate", "Basic realm=Restricted")
		http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
	}
}

func basicAuthHandler(h http.Handler, login, password []byte) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if len(login) == 0 || len(password) == 0 {
			h.ServeHTTP(w, r)
			return
		}

		const basicAuthPrefix string = "Basic "

		auth := r.Header.Get("Authorization")
		if strings.HasPrefix(auth, basicAuthPrefix) {
			payload, err := base64.StdEncoding.DecodeString(auth[len(basicAuthPrefix):])
			if err == nil {
				pair := bytes.SplitN(payload, []byte(":"), 2)
				if len(pair) == 2 && bytes.Equal(pair[0], login) && bytes.Equal(pair[1], password) {
					// Delegate request to the given handle
					h.ServeHTTP(w, r)
					return
				}
			}
		}

		w.Header().Set("WWW-Authenticate", "Basic realm=Restricted")
		http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
	})
}

func getRouter(wsHandler http.Handler, apiServer *APIServer, login string, password string) *httprouter.Router {
	lbytes := []byte(login)
	pbytes := []byte(password)

	router := httprouter.New()

	router.Handler("GET", "/wg", basicAuthHandler(wsHandler, lbytes, pbytes))

	router.Handle("GET", "/peers", basicAuth(apiServer.peers, lbytes, pbytes))
	router.Handle("GET", "/peer/:peer", basicAuth(apiServer.peerInterfaces, lbytes, pbytes))

	router.Handle("POST", "/all", basicAuth(apiServer.putInterfacesAll, lbytes, pbytes))
	router.Handle("POST", "/all/delete", basicAuth(apiServer.delInterfacesAll, lbytes, pbytes))
	router.Handle("POST", "/all/sync", basicAuth(apiServer.syncPeerAll, lbytes, pbytes))

	router.Handle("POST", "/peer/:peer", basicAuth(apiServer.putInterfaces, lbytes, pbytes))
	router.Handle("POST", "/peer/:peer/delete", basicAuth(apiServer.delInterfaces, lbytes, pbytes))
	router.Handle("POST", "/peer/:peer/sync", basicAuth(apiServer.syncPeer, lbytes, pbytes))

	return router
}
