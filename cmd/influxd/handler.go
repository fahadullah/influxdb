package main

import (
	"net/http"
	"strings"

	"github.com/influxdb/influxdb"
	"github.com/influxdb/influxdb/httpd"
	"github.com/influxdb/influxdb/messaging"
	"github.com/influxdb/influxdb/raft"
)

// Handler represents an HTTP handler for InfluxDB node.
// Depending on its role, it will serve many different endpoints.
type Handler struct {
	Log    *raft.Log
	Broker *influxdb.Broker
	Server *influxdb.Server
	Config *Config
}

// NewHandler returns a new instance of Handler.
func NewHandler() *Handler {
	return &Handler{}
}

// ServeHTTP responds to HTTP request to the handler.
func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if strings.HasPrefix(r.URL.Path, "/raft") {
		h.serveRaft(w, r)
		return
	}
	if strings.HasPrefix(r.URL.Path, "/messaging") {
		h.serveMessaging(w, r)
		return
	}

	h.serveData(w, r)
}

func (h *Handler) serveMessaging(w http.ResponseWriter, r *http.Request) {
	if h.Broker != nil {
		mh := &messaging.Handler{
			Broker:      h.Broker.Broker,
			RaftHandler: &raft.Handler{Log: h.Log},
		}
		mh.ServeHTTP(w, r)
		return
	}

	b := h.Server.BrokerURLs()
	http.Redirect(w, r, b[0].String(), http.StatusMovedPermanently)
}

// serveRaft responds to raft requests.
func (h *Handler) serveRaft(w http.ResponseWriter, r *http.Request) {
	if h.Log != nil {
		rh := raft.Handler{Log: h.Log}
		rh.ServeHTTP(w, r)
		return
	}

	// TODO: Redirect to broker.
}

func (h *Handler) serveData(w http.ResponseWriter, r *http.Request) {
	if h.Server != nil {
		sh := httpd.NewHandler(h.Server, h.Config.Authentication.Enabled, version)
		sh.WriteTrace = h.Config.Logging.WriteTracing
		sh.ServeHTTP(w, r)
		return
	}
}
