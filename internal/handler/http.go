package handler

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"

	"strata-weave/internal/model"
	"strata-weave/internal/service"
)

type Handler struct{ app *service.Service }

func New(app *service.Service) http.Handler {
	h := &Handler{app: app}
	m := http.NewServeMux()
	m.HandleFunc("/", h.home)
	m.HandleFunc("/api/dashboard", h.dashboard)
	m.HandleFunc("/api/trenches", h.trenches)
	m.HandleFunc("/api/units", h.units)
	m.HandleFunc("/api/relations", h.relations)
	m.HandleFunc("/api/finds", h.finds)
	m.HandleFunc("/api/samples/", h.samples)
	m.HandleFunc("/api/records", h.records)
	m.HandleFunc("/api/observations", h.observations)
	m.HandleFunc("/api/alerts", h.alerts)
	return m
}
func write(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}
func errStatus(e error) int {
	if errors.Is(e, model.ErrNotFound) {
		return 404
	}
	if errors.Is(e, model.ErrInvalidInput) || errors.Is(e, model.ErrInvalidState) || errors.Is(e, model.ErrCycle) || errors.Is(e, model.ErrCrossTrench) || errors.Is(e, model.ErrUnreviewedFind) {
		return 400
	}
	return 500
}
func decode(r *http.Request, v any) error { return json.NewDecoder(r.Body).Decode(v) }
func (h *Handler) home(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	page := `<!doctype html><html><head><meta charset="utf-8"><title>Strata Weave</title><style>body{font:16px system-ui;margin:2rem;max-width:900px}header{display:flex;justify-content:space-between}main{display:grid;grid-template-columns:repeat(5,1fr);gap:1rem}.metric{border:1px solid #bbb;padding:1rem;border-radius:6px}.metric b{display:block;font-size:2rem}button{padding:.5rem}</style></head><body><header><h1>Strata Weave</h1><button onclick="load()">Refresh</button></header><p>Archaeological field operations</p><main id="metrics"></main><script>async function load(){let d=await fetch('/api/dashboard').then(r=>r.json());metrics.innerHTML=Object.entries(d).map(function(pair){return '<div class=metric><b>'+pair[1]+'</b>'+pair[0]+'</div>'}).join('')}load()</script></body></html>`
	_, _ = w.Write([]byte(page))
}
func (h *Handler) dashboard(w http.ResponseWriter, r *http.Request) {
	d, e := h.app.Dashboard(r.Context())
	if e != nil {
		write(w, 500, map[string]string{"error": e.Error()})
		return
	}
	write(w, 200, d)
}
func (h *Handler) trenches(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		v, e := h.app.ListTrenches(r.Context())
		if e != nil {
			write(w, 500, map[string]string{"error": e.Error()})
			return
		}
		write(w, 200, v)
		return
	}
	if r.Method == "POST" {
		var t model.Trench
		if e := decode(r, &t); e != nil {
			write(w, 400, map[string]string{"error": e.Error()})
			return
		}
		v, e := h.app.CreateTrench(r.Context(), t)
		if e != nil {
			write(w, errStatus(e), map[string]string{"error": e.Error()})
			return
		}
		write(w, 201, v)
		return
	}
	w.WriteHeader(405)
}
func (h *Handler) units(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		v, e := h.app.ListUnits(r.Context(), model.UnitFilter{TrenchID: r.URL.Query().Get("trench_id"), Phase: r.URL.Query().Get("phase")})
		if e != nil {
			write(w, 500, map[string]string{"error": e.Error()})
			return
		}
		write(w, 200, v)
		return
	}
	if r.Method == "POST" {
		var u model.Unit
		if e := decode(r, &u); e != nil {
			write(w, 400, map[string]string{"error": e.Error()})
			return
		}
		v, e := h.app.CreateUnit(r.Context(), u)
		if e != nil {
			write(w, errStatus(e), map[string]string{"error": e.Error()})
			return
		}
		write(w, 201, v)
		return
	}
	w.WriteHeader(405)
}
func (h *Handler) relations(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		v, e := h.app.Matrix(r.Context())
		if e != nil {
			write(w, 500, map[string]string{"error": e.Error()})
			return
		}
		write(w, 200, v)
		return
	}
	if r.Method == "POST" {
		var v model.Relation
		if e := decode(r, &v); e != nil {
			write(w, 400, map[string]string{"error": e.Error()})
			return
		}
		x, e := h.app.AddStratigraphicRelation(r.Context(), v)
		if e != nil {
			write(w, errStatus(e), map[string]string{"error": e.Error()})
			return
		}
		write(w, 201, x)
		return
	}
	w.WriteHeader(405)
}
func (h *Handler) finds(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		w.WriteHeader(405)
		return
	}
	var v model.Find
	if e := decode(r, &v); e != nil {
		write(w, 400, map[string]string{"error": e.Error()})
		return
	}
	x, e := h.app.CreateFind(r.Context(), v)
	if e != nil {
		write(w, errStatus(e), map[string]string{"error": e.Error()})
		return
	}
	write(w, 201, x)
}
func (h *Handler) samples(w http.ResponseWriter, r *http.Request) {
	parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if len(parts) < 3 {
		w.WriteHeader(404)
		return
	}
	id := parts[2]
	if r.Method == "POST" && len(parts) == 3 {
		var v model.Sample
		if e := decode(r, &v); e != nil {
			write(w, 400, map[string]string{"error": e.Error()})
			return
		}
		x, e := h.app.CreateSample(r.Context(), v)
		if e != nil {
			write(w, errStatus(e), map[string]string{"error": e.Error()})
			return
		}
		write(w, 201, x)
		return
	}
	if r.Method == "POST" && len(parts) == 4 && parts[3] == "dispatch" {
		var p struct {
			Lab string `json:"lab"`
		}
		decode(r, &p)
		if e := h.app.DispatchSample(r.Context(), id, p.Lab); e != nil {
			write(w, errStatus(e), map[string]string{"error": e.Error()})
			return
		}
		write(w, 200, map[string]string{"status": "dispatched"})
		return
	}
	w.WriteHeader(405)
}
func (h *Handler) records(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		v, e := h.app.ListRecords(r.Context(), model.RecordFilter{Status: r.URL.Query().Get("status"), UnitID: r.URL.Query().Get("unit_id")})
		if e != nil {
			write(w, 500, map[string]string{"error": e.Error()})
			return
		}
		write(w, 200, v)
		return
	}
	if r.Method == "POST" {
		var v model.FieldRecord
		if e := decode(r, &v); e != nil {
			write(w, 400, map[string]string{"error": e.Error()})
			return
		}
		x, e := h.app.CreateRecord(r.Context(), v)
		if e != nil {
			write(w, errStatus(e), map[string]string{"error": e.Error()})
			return
		}
		write(w, 201, x)
		return
	}
	w.WriteHeader(405)
}
func (h *Handler) observations(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		v, e := h.app.ListObservations(r.Context(), model.ObservationFilter{UnitID: r.URL.Query().Get("unit_id"), Metric: r.URL.Query().Get("metric")})
		if e != nil {
			write(w, 500, map[string]string{"error": e.Error()})
			return
		}
		write(w, 200, v)
		return
	}
	if r.Method == "POST" {
		var v model.Observation
		if e := decode(r, &v); e != nil {
			write(w, 400, map[string]string{"error": e.Error()})
			return
		}
		if v.At.IsZero() {
			v.At = time.Now()
		}
		e := h.app.IngestObservation(context.Background(), v)
		if e != nil {
			write(w, errStatus(e), map[string]string{"error": e.Error()})
			return
		}
		write(w, 202, v)
		return
	}
	w.WriteHeader(405)
}
func (h *Handler) alerts(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		v, e := h.app.ListAlerts(r.Context(), model.AlertFilter{Status: r.URL.Query().Get("status"), UnitID: r.URL.Query().Get("unit_id")})
		if e != nil {
			write(w, 500, map[string]string{"error": e.Error()})
			return
		}
		write(w, 200, v)
		return
	}
	if r.Method == "POST" {
		var v model.Alert
		if e := decode(r, &v); e != nil {
			write(w, 400, map[string]string{"error": e.Error()})
			return
		}
		x, e := h.app.CreateAlert(r.Context(), v)
		if e != nil {
			write(w, errStatus(e), map[string]string{"error": e.Error()})
			return
		}
		write(w, 201, x)
		return
	}
	w.WriteHeader(405)
}
