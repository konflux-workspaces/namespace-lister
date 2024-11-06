package acceptance

import (
	"cmp"
	"fmt"
	"os"

	"github.com/cucumber/godog"
	"github.com/konflux-workspaces/namespace-lister/acceptance/pkg/rest"
	"golang.org/x/net/context"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func InjectSteps(ctx *godog.ScenarioContext) {
	ctx.Step(`^the user can retrieve only the namespaces they have access to$`, TheUserCanRetrieveOnlyTheNamespacesTheyHaveAccessTo)
	ctx.Step(`^user has access to "([^"]*)" namespaces$`, UserHasAccessToNNamespaces)
}

func UserHasAccessToNNamespaces(ctx context.Context, number int) (context.Context, error) {
	cli, err := rest.BuildDefaultHostClient()
	if err != nil {
		return ctx, err
	}

	// create serviceaccount
	if err := cli.Create(ctx, &corev1.ServiceAccount{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "user",
			Namespace: "default",
		},
	}); err != nil && !errors.IsAlreadyExists(err) {
		return ctx, err
	}

	// create namespaces
	for i := range number {
		if err := cli.Create(ctx, &corev1.Namespace{
			ObjectMeta: metav1.ObjectMeta{
				Name: fmt.Sprintf("run-1-%d", i),
				Labels: map[string]string{
					"namespace-lister/scope": "acceptance-tests",
				},
			},
		}); err != nil {
			return ctx, err
		}

		cli.Create(ctx, &rbacv1.RoleBinding{
			ObjectMeta: metav1.ObjectMeta{
				Name:      fmt.Sprintf("run-1-%d", i),
				Namespace: fmt.Sprintf("run-1-%d", i),
			},
			RoleRef: rbacv1.RoleRef{
				Kind:     "ClusterRole",
				Name:     "namespace-get",
				APIGroup: rbacv1.GroupName,
			},
			Subjects: []rbacv1.Subject{
				{
					Kind:     "User",
					APIGroup: rbacv1.GroupName,
					Name:     "user",
				},
			},
		})
	}

	return ctx, nil
}

func TheUserCanRetrieveOnlyTheNamespacesTheyHaveAccessTo(ctx context.Context) (context.Context, error) {
	// build impersonating client
	cfg, err := rest.NewDefaultClientConfig()
	if err != nil {
		return ctx, err
	}
	cfg.Impersonate.UserName = "user"
	cfg.Host = cmp.Or(os.Getenv("KONFLUX_ADDRESS"), "https://localhost:10443")
	cli, err := rest.BuildClient(cfg)
	if err != nil {
		return ctx, err
	}

	nn := corev1.NamespaceList{}
	if err := cli.List(ctx, &nn); err != nil {
		return ctx, err
	}

	if lnni := len(nn.Items); lnni != 10 {
		return ctx, fmt.Errorf("expected 10 namespaces, found %d", lnni)
	}

	return ctx, nil
}
