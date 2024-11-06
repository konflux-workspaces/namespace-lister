package main

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	openshiftapiv1 "github.com/openshift/api/project/v1"
	kerrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/rest"
)

var _ http.Handler = &getNamespaceHandler{}

type getNamespaceHandler struct {
	cfg *rest.Config
	log *slog.Logger
}

func newGetNamespaceHandler(cfg *rest.Config, log *slog.Logger) http.Handler {
	return &getNamespaceHandler{
		cfg: cfg,
		log: log,
	}
}

func (h *getNamespaceHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// build client
	h.log.Info("serving get")
	cli, err := buildImpersonatingClient(h.cfg, r.Header.Get("X-Username"))
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}

	// fetch project
	project := openshiftapiv1.Project{}
	name := r.PathValue("name")
	if err := cli.Get(r.Context(), types.NamespacedName{Name: name}, &project); err != nil {
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

	// map back to namespace
	namespace := convertProjectToNamespace(&project)
	encoded, err := json.Marshal(&namespace)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}

	w.Write(encoded)
}
