package main

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
)

type Handler struct {
	store     *Store
	simulator *Simulator
}

func NewHandler(store *Store, simulator *Simulator) *Handler {
	return &Handler{
		store:     store,
		simulator: simulator,
	}
}

func (h *Handler) Register(r chi.Router) {
	r.Post("/webhooks/orders", h.HandleWebhook)
	r.Get("/healthz", h.HandleHealthz)
	r.Get("/orders", h.HandleListOrders)
}

func (h *Handler) HandleWebhook(w http.ResponseWriter, r *http.Request) {
	var payload OrderPayload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if !h.store.TryInsert(payload) { // false = already exists
		w.WriteHeader(http.StatusAccepted)
		return
	}

	go h.simulator.Run(payload.OrderID) //nolint:contextcheck // run asynchronously

	w.WriteHeader(http.StatusAccepted)
}

func (h *Handler) HandleHealthz(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
}

func (h *Handler) HandleListOrders(w http.ResponseWriter, _ *http.Request) {
	orders := h.store.List()
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(orders); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	}
}
