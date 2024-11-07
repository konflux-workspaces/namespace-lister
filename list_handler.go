package main

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	kerrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/client-go/rest"
)

var _ http.Handler = &listNamespacesHandler{}

type listNamespacesHandler struct {
	cfg        *rest.Config
	log        *slog.Logger
	cache      *Cache
	userHeader string
}

func newListNamespacesHandler(cfg *rest.Config, log *slog.Logger, cache *Cache, userHeader string) http.Handler {
	return &listNamespacesHandler{
		cfg:        cfg,
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
			w.Write([]byte(serr.Error()))
			return
		}

		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}

	// build response
	// for PoC limited to JSON
	b, err := json.Marshal(nn)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}

	w.Header().Add(HttpContentType, HttpContentTypeApplication)
	w.Write(b)
}
