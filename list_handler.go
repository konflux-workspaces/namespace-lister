package main

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	corev1 "k8s.io/api/core/v1"
	kerrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/rest"
)

var _ http.Handler = &listNamespacesHandler{}

type listNamespacesHandler struct {
	cfg        *rest.Config
	log        *slog.Logger
	ctrl       *Controller
	userHeader string
}

func newListNamespacesHandler(cfg *rest.Config, log *slog.Logger, ctrl *Controller, userHeader string) http.Handler {
	return &listNamespacesHandler{
		cfg:        cfg,
		log:        log,
		ctrl:       ctrl,
		userHeader: userHeader,
	}
}

func (h *listNamespacesHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.log.Info("received list request")
	// retrieve projects as the user
	nn, err := h.ctrl.ListNamespaces(r.Context(), r.Header.Get(h.userHeader))
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
	l := corev1.NamespaceList{
		TypeMeta: metav1.TypeMeta{Kind: "NamespaceList", APIVersion: "v1"},
		Items:    nn,
	}
	b, err := json.Marshal(l)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}

	w.Header().Add("Content-Type", "application/json;charset=utf-8")
	w.Write(b)
}
