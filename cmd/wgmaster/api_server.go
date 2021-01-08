package main

import (
	"encoding/json"
	"fmt"
	"net"
	"net/http"

	"github.com/pnforge/wgadmin/wglib/master"
	"github.com/julienschmidt/httprouter"
	"go.uber.org/zap"
)

type APIServer struct {
	Md     *master.Master
	logger *zap.Logger
}

type ErrResponse struct {
	Error string `json:"error"`
}

func (a *APIServer) peers(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	peers := a.Md.Peers()

	a.writeJSON(w, peers)
}

func (a *APIServer) delInterfaces(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	peer := ps.ByName("peer")
	if ip := net.ParseIP(peer); ip == nil {
		a.badRequest(w, r, ps, fmt.Sprintf("'%s' is not valid peer address", peer))
		return
	}

	decoder := json.NewDecoder(r.Body)
	keys := []string{}

	if err := decoder.Decode(&keys); err != nil {
		a.badRequest(w, r, ps, "failed to decode interface keys: "+err.Error())
		return
	}

	if err := a.Md.DelInterfacesPeer(peer, keys); err != nil {
		msg := ErrResponse{Error: "request failed: " + err.Error()}
		a.writeJSON(w, msg)

		return
	}

	a.writeJSON(w, "OK")
}

func (a *APIServer) delInterfacesAll(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	decoder := json.NewDecoder(r.Body)
	keys := []string{}

	if err := decoder.Decode(&keys); err != nil {
		a.badRequest(w, r, ps, "failed to decode interface keys: "+err.Error())
		return
	}

	results := make(map[string]string)

	for _, peer := range a.Md.Peers() {
		if err := a.Md.DelInterfacesPeer(peer, keys); err != nil {
			results[peer] = err.Error()
		} else {
			results[peer] = "ok"
		}
	}

	a.writeJSON(w, map[string]interface{}{"results": results})
}

func (a *APIServer) syncPeer(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	peer := ps.ByName("peer")
	if ip := net.ParseIP(peer); ip == nil {
		a.badRequest(w, r, ps, fmt.Sprintf("'%s' is not valid peer address", peer))
		return
	}

	decoder := json.NewDecoder(r.Body)
	ifs := []master.Interface{}

	if err := decoder.Decode(&ifs); err != nil {
		a.badRequest(w, r, ps, "failed to decode interfaces: "+err.Error())
		return
	}

	if err := a.Md.PutInterfacesPeer(peer, ifs, true); err != nil {
		msg := ErrResponse{Error: "request failed: " + err.Error()}
		a.writeJSON(w, msg)

		return
	}

	a.writeJSON(w, "OK")
}

func (a *APIServer) syncPeerAll(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	decoder := json.NewDecoder(r.Body)
	ifs := []master.Interface{}

	if err := decoder.Decode(&ifs); err != nil {
		a.badRequest(w, r, ps, "failed to decode interfaces: "+err.Error())
		return
	}

	results := make(map[string]string)

	for _, peer := range a.Md.Peers() {
		if err := a.Md.PutInterfacesPeer(peer, ifs, true); err != nil {
			results[peer] = err.Error()
		} else {
			results[peer] = "ok"
		}
	}

	a.writeJSON(w, map[string]interface{}{"results": results})
}

func (a *APIServer) putInterfaces(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	peer := ps.ByName("peer")
	if ip := net.ParseIP(peer); ip == nil {
		a.badRequest(w, r, ps, fmt.Sprintf("'%s' is not valid peer address", peer))
		return
	}

	decoder := json.NewDecoder(r.Body)
	ifs := []master.Interface{}

	if err := decoder.Decode(&ifs); err != nil {
		a.badRequest(w, r, ps, "failed to decode interfaces: "+err.Error())
		return
	}

	if err := a.Md.PutInterfacesPeer(peer, ifs, false); err != nil {
		msg := ErrResponse{Error: "request failed: " + err.Error()}
		a.writeJSON(w, msg)

		return
	}

	a.writeJSON(w, "OK")
}

func (a *APIServer) putInterfacesAll(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	decoder := json.NewDecoder(r.Body)
	ifs := []master.Interface{}

	if err := decoder.Decode(&ifs); err != nil {
		a.badRequest(w, r, ps, "failed to decode interfaces: "+err.Error())
		return
	}

	results := make(map[string]string)

	for _, peer := range a.Md.Peers() {
		if err := a.Md.PutInterfacesPeer(peer, ifs, false); err != nil {
			results[peer] = err.Error()
		} else {
			results[peer] = "ok"
		}
	}

	a.writeJSON(w, map[string]interface{}{"results": results})
}

func (a *APIServer) peerInterfaces(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	peer := ps.ByName("peer")
	if ip := net.ParseIP(peer); ip == nil {
		a.badRequest(w, r, ps, fmt.Sprintf("'%s' is not valid peer address", peer))
		return
	}

	interfaces, err := a.Md.PeerInterfaces(peer)
	if err != nil {
		a.logger.Info("peerInterfaces: error response", zap.String("peer", peer),
			zap.String("remote", r.RemoteAddr), zap.String("error", err.Error()))
		a.writeJSON(w, ErrResponse{Error: err.Error()})

		return
	}

	a.writeJSON(w, struct {
		Interfaces []master.Interface `json:"peers"`
	}{Interfaces: interfaces})
}

func (a *APIServer) badRequest(w http.ResponseWriter, r *http.Request, ps httprouter.Params, message string) {
	a.logger.Info("bad request", zap.String("message", message), zap.String("remote", r.RemoteAddr),
		zap.String("request", r.RequestURI), zap.String("params", fmt.Sprintf("%+v", ps)))

	msg := ErrResponse{Error: fmt.Sprintf("bad request: %s", message)}

	w.WriteHeader(http.StatusBadRequest)
	w.Header().Set("Content-Type", "application/json")

	if err := json.NewEncoder(w).Encode(msg); err != nil {
		a.logger.Error("badRequest: failed to encode json data", zap.Error(err))
	}
}

func (a *APIServer) writeJSON(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")

	if err := json.NewEncoder(w).Encode(data); err != nil {
		a.logger.Error("writeJson: failed to encode json data", zap.Error(err))
	}
}
