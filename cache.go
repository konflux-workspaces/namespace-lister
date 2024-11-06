package main

import (
	"context"
	"fmt"
	"log/slog"

	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apiserver/pkg/authentication/user"
	"k8s.io/apiserver/pkg/authorization/authorizer"

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/cache"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type Cache struct {
	client.Reader
	l *slog.Logger
}

func NewCache(ctx context.Context, l *slog.Logger) (*Cache, error) {
	cfg := ctrl.GetConfigOrDie()

	s := runtime.NewScheme()
	if err := corev1.AddToScheme(s); err != nil {
		return nil, err
	}
	if err := rbacv1.AddToScheme(s); err != nil {
		return nil, err
	}
	oo := []client.Object{
		&corev1.Namespace{},
		&rbacv1.RoleBinding{},
		&rbacv1.ClusterRole{},
		&rbacv1.ClusterRoleBinding{},
		&rbacv1.Role{},
	}
	c, err := cache.New(cfg, cache.Options{
		Scheme: s,
		ByObject: map[client.Object]cache.ByObject{
			&corev1.Namespace{}:          {},
			&rbacv1.RoleBinding{}:        {},
			&rbacv1.ClusterRole{}:        {},
			&rbacv1.ClusterRoleBinding{}: {},
			&rbacv1.Role{}:               {},
		},
	})
	if err != nil {
		return nil, err
	}

	for _, o := range oo {
		_, err := c.GetInformer(ctx, o)
		if err != nil {
			return nil, fmt.Errorf("error starting cache: getting informer for %s: %w", o.GetObjectKind().GroupVersionKind().String(), err)
		}
	}

	go func() {
		if err := c.Start(ctx); err != nil {
			panic(err)
		}
	}()
	if !c.WaitForCacheSync(ctx) {
		return nil, fmt.Errorf("error starting the cache")
	}

	return &Cache{Reader: c, l: l}, nil
}

func (c *Cache) ListNamespaces(ctx context.Context, username string) (*corev1.NamespaceList, error) {
	// list role bindings
	nn := v1.NamespaceList{}
	if err := c.List(ctx, &nn); err != nil {
		return nil, err
	}

	auz := NewAuthorizer(ctx, c, c.l)
	rnn := []corev1.Namespace{}
	for _, ns := range nn.Items {
		d, _, err := auz.Authorize(ctx, authorizer.AttributesRecord{
			User:            &user.DefaultInfo{Name: username},
			Verb:            "get",
			Resource:        "namespaces",
			APIGroup:        "",
			APIVersion:      "v1",
			Name:            ns.Name,
			Namespace:       ns.Name,
			ResourceRequest: true,
		})
		if err != nil {
			return nil, err
		}

		c.l.Info("evaluated user access to namespace", "namespace", ns.Name, "user", username, "decision", d)
		if d == authorizer.DecisionAllow {
			rnn = append(rnn, ns)
		}
	}
	nn.Items = rnn

	return &nn, nil
}
