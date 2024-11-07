package main

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	kerrors "k8s.io/apimachinery/pkg/api/errors"
)

var _ http.Handler = &listNamespacesHandler{}

type listNamespacesHandler struct {
	log        *slog.Logger
	cache      *Cache
	userHeader string
}

func newListNamespacesHandler(log *slog.Logger, cache *Cache, userHeader string) http.Handler {
	return &listNamespacesHandler{
		log:        log,
		cache:      cache,
		userHeader: userHeader,
	}
}

func (h *listNamespacesHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.log.Info("received list request")
	// retrieve projects as the user
	nn, err := h.cache.ListNamespaces(r.Context(), r.Header.Get(h.userHeader))
	if err != nil {
		serr := &kerrors.StatusError{}
		if errors.As(err, &serr) {
			w.WriteHeader(int(serr.Status().Code))
			h.write(w, []byte(serr.Error()))
			return
		}

		w.WriteHeader(http.StatusInternalServerError)
		h.write(w, []byte(err.Error()))
		return
	}

	// build response
	// for PoC limited to JSON
	b, err := json.Marshal(nn)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		h.write(w, []byte(err.Error()))
		return
	}

	w.Header().Add(HttpContentType, HttpContentTypeApplication)
	h.write(w, b)
}

func (h *listNamespacesHandler) write(w http.ResponseWriter, data []byte) bool {
	if _, err := w.Write(data); err != nil {
		h.log.Error("error writing reply", "error", err)
		w.WriteHeader(http.StatusInternalServerError)
		return false
	}
	return true
}
