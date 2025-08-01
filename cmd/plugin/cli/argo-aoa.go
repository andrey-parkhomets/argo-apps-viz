package cli

import (
	"context"
	_ "embed"

	"argo-apps-viz/pkg/logger"
	"argo-apps-viz/pkg/model/appsofapps"

	"github.com/argoproj/argo-cd/v3/pkg/client/clientset/versioned/typed/application/v1alpha1"
	"github.com/go-echarts/go-echarts/v2/components"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var (
	appsOfAppsFile = "apps-of-apps.html"
)

var argoAppsOffApps = &cobra.Command{
	Use:   "apps-of-apps",
	Short: "Generate documentation from your ArgoCD applications and applicationsSets within your cluster",
	RunE: func(c *cobra.Command, args []string) error {
		logger := logger.NewLogger()
		logger.Info("Generating Dependency Chart :)")
		graph, err := runAoa(c.Flags())
		if err != nil {
			return err
		}
		err = CreateFile(appsOfAppsFile, graph)
		if err != nil {
			return err
		}
		logger.Info("Finished look in: " + appsOfAppsFile)
		return err
	},

	SilenceUsage: true,
}

func init() {
	rootCmd.AddCommand(argoAppsOffApps)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// graphAoaCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	argoAppsOffApps.Flags().StringArray("start", []string{}, "Application where to start a graph. Useful when just printing a subgraph")
	argoAppsOffApps.Flags().StringArray("stop", []string{}, "When using a Recursive apps of apps pattern, or you want to stop at a specific application at this")
	argoAppsOffApps.Flags().BoolP("tree", "t", false, "Set this if you wand a tree instead of a graph")
}

func runAoa(flags *pflag.FlagSet) (components.Charter, error) {
	log := logger.NewLogger()
	config, err := KubernetesConfigFlags.ToRESTConfig()
	if err != nil {
		log.Error(err)
		return nil, err
	}

	argoclient, err := v1alpha1.NewForConfig(config)
	if err != nil {
		log.Error(err)
		return nil, err
	}

	// Get namespace from flags
	namespace, err := flags.GetString("namespace")
	if err != nil {
		namespace = "argocd" // fallback to default
	}

	var ctx = context.Background()
	applicationList, err := argoclient.Applications(namespace).List(ctx, v1.ListOptions{})
	if err != nil {
		log.Info("Problem while getting ArgoCd domains")
		log.Error(err)
		return nil, err
	}

	applicationSetList, err := argoclient.ApplicationSets(namespace).List(ctx, v1.ListOptions{})
	if err != nil {
		log.Info("Problem while getting ArgoCd domains")
		log.Error(err)
		return nil, err
	}

	isTree, err := flags.GetBool("tree")
	if err != nil {
		log.Error(err)
		return nil, err
	}
	starts, err := flags.GetStringArray("start")
	if err != nil {
		log.Error(err)
		return nil, err
	}
	if len(starts) != 0 {
		log.Info("Using applications as starting points of graph:")
		for i, start := range starts {
			log.Info("  %d, %s", i, start)
		}
	}
	stops, err := flags.GetStringArray("stop")
	if err != nil {
		log.Error(err)
		return nil, err
	}
	if len(stops) != 0 {
		log.Info("Using applications as stopping points of graph:")
		for i, stop := range stops {
			log.Info("  %d, %s", i, stop)
		}
	}

	if isTree {
		return appsofapps.RenderTree(applicationSetList, applicationList, stops), nil
	} else {
		return appsofapps.AppsOfAppsRenderGraph(applicationSetList, applicationList, starts, stops), nil
	}
}
