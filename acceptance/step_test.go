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
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func InjectSteps(ctx *godog.ScenarioContext) {
	//read
	ctx.Step(`^user has access to a namespace$`,
		func(ctx context.Context) (context.Context, error) { return UserHasAccessToNNamespaces(ctx, 1) })
	ctx.Step(`^the user can retrieve the namespace$`, TheUserCanRetrieveTheNamespace)

	// list
	ctx.Step(`^user has access to "([^"]*)" namespaces$`, UserHasAccessToNNamespaces)
	ctx.Step(`^the user can retrieve only the namespaces they have access to$`, TheUserCanRetrieveOnlyTheNamespacesTheyHaveAccessTo)
}

func UserHasAccessToNNamespaces(ctx context.Context, number int) (context.Context, error) {
	run := ctx.Value("run").(string)

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
				Name: fmt.Sprintf("run-%s-%d", run, i),
				Labels: map[string]string{
					"namespace-lister/scope":    "acceptance-tests",
					"namespace-lister/test-run": run,
				},
			},
		}); err != nil {
			return ctx, err
		}

		cli.Create(ctx, &rbacv1.RoleBinding{
			ObjectMeta: metav1.ObjectMeta{
				Name:      fmt.Sprintf("run-%s-%d", run, i),
				Namespace: fmt.Sprintf("run-%s-%d", run, i),
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
	cli, err := buildUserClient()
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

func TheUserCanRetrieveTheNamespace(ctx context.Context) (context.Context, error) {
	run := ctx.Value("run").(string)

	cli, err := buildUserClient()
	if err != nil {
		return ctx, err
	}

	n := corev1.Namespace{}
	k := types.NamespacedName{Name: fmt.Sprintf("run-%s-0", run)}
	if err := cli.Get(ctx, k, &n); err != nil {
		return ctx, err
	}

	return ctx, nil
}

func buildUserClient() (client.Client, error) {
	// build impersonating client
	cfg, err := rest.NewDefaultClientConfig()
	if err != nil {
		return nil, err
	}
	cfg.Impersonate.UserName = "user"
	cfg.Host = cmp.Or(os.Getenv("KONFLUX_ADDRESS"), "https://localhost:10443")
	return rest.BuildClient(cfg)
}
