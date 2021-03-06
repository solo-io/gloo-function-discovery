package main

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/solo-io/gloo-storage/crd"
	"github.com/spf13/cobra"

	"github.com/solo-io/gloo-function-discovery/internal/eventloop"
	"github.com/solo-io/gloo-function-discovery/internal/options"
	"github.com/solo-io/gloo/pkg/bootstrap"
	"github.com/solo-io/gloo/pkg/log"
	"github.com/solo-io/gloo/pkg/signals"
)

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

var (
	opts          bootstrap.Options
	discoveryOpts options.DiscoveryOptions
)

var rootCmd = &cobra.Command{
	Use:   "gloo-function-discovery",
	Short: "discovers functions for swagger, google functions, and lambda upstreams",

	RunE: func(cmd *cobra.Command, args []string) error {
		stop := signals.SetupSignalHandler()
		errs := make(chan error)

		finished := make(chan error)
		go func() { finished <- eventloop.Run(opts, discoveryOpts, stop, errs) }()
		go func() {
			for {
				select {
				case err := <-errs:
					log.Warnf("discovery error: %v", err)
				}
			}
		}()
		return <-finished
	},
}

func init() {
	// config watcher
	rootCmd.PersistentFlags().StringVar(&opts.ConfigWatcherOptions.Type, "storage.type", bootstrap.WatcherTypeKube, fmt.Sprintf("storage backend for config objects. supported: [%s]", strings.Join(bootstrap.SupportedCwTypes, " | ")))
	rootCmd.PersistentFlags().DurationVar(&opts.ConfigWatcherOptions.SyncFrequency, "storage.refreshrate", time.Second, "refresh rate for polling config")

	// secret watcher
	rootCmd.PersistentFlags().StringVar(&opts.SecretWatcherOptions.Type, "secrets.type", bootstrap.WatcherTypeKube, fmt.Sprintf("storage backend for secrets. supported: [%s]", strings.Join(bootstrap.SupportedSwTypes, " | ")))
	rootCmd.PersistentFlags().DurationVar(&opts.SecretWatcherOptions.SyncFrequency, "secrets.refreshrate", time.Second, "refresh rate for polling secrets")

	// file watcher
	rootCmd.PersistentFlags().StringVar(&opts.FileWatcherOptions.Type, "files.type", bootstrap.WatcherTypeKube, fmt.Sprintf("storage backend for files. supported: [%s]", strings.Join(bootstrap.SupportedFwTypes, " | ")))
	rootCmd.PersistentFlags().DurationVar(&opts.FileWatcherOptions.SyncFrequency, "files.refreshrate", time.Second, "refresh rate for polling files")

	// file
	rootCmd.PersistentFlags().StringVar(&opts.FileOptions.ConfigDir, "file.config.dir", "_gloo_config", "root directory to use for storing gloo config files")
	rootCmd.PersistentFlags().StringVar(&opts.FileOptions.FilesDir, "file.files.dir", "_gloo_config", "root directory to use for storing input files")
	rootCmd.PersistentFlags().StringVar(&opts.FileOptions.SecretDir, "file.secret.dir", "_gloo_secrets", "root directory to use for storing gloo secret files")

	// kube
	rootCmd.PersistentFlags().StringVar(&opts.KubeOptions.MasterURL, "master", "", "url of the kubernetes apiserver. not needed if running in-cluster")
	rootCmd.PersistentFlags().StringVar(&opts.KubeOptions.KubeConfig, "kubeconfig", "", "path to kubeconfig file. not needed if running in-cluster")
	rootCmd.PersistentFlags().StringVar(&opts.KubeOptions.Namespace, "kube.namespace", crd.GlooDefaultNamespace, "namespace to read/write gloo storage objects")

	// consul
	rootCmd.PersistentFlags().StringVar(&opts.ConsulOptions.RootPath, "consul.root", "gloo", "prefix for all k/v pairs stored in consul by gloo, when using consul for storage")
	rootCmd.PersistentFlags().StringVar(&opts.ConsulOptions.Datacenter, "consul.datacenter", "", "datacenter of the consul server when using consul for storage or service discovery")
	rootCmd.PersistentFlags().StringVar(&opts.ConsulOptions.Address, "consul.address", "", "address (including port) of the consul server to connect to when using consul for storage")
	rootCmd.PersistentFlags().StringVar(&opts.ConsulOptions.Scheme, "consul.scheme", "", "uri scheme for the consul server")
	rootCmd.PersistentFlags().StringVar(&opts.ConsulOptions.Token, "consul.token", "", "token is used to provide a per-request ACL token to override the default")
	rootCmd.PersistentFlags().StringVar(&opts.ConsulOptions.Username, "consul.username", "", "username for authenticating to the consul server, if using basic auth")
	rootCmd.PersistentFlags().StringVar(&opts.ConsulOptions.Password, "consul.password", "", "password for authenticating to the consul server, if using basic auth")

	// vault
	rootCmd.PersistentFlags().StringVar(&opts.VaultOptions.VaultAddr, "vault.addr", "", "url for vault server")
	rootCmd.PersistentFlags().StringVar(&opts.VaultOptions.AuthToken, "vault.token", "", "auth token for reading vault secrets")
	rootCmd.PersistentFlags().IntVar(&opts.VaultOptions.Retries, "vault.retries", 3, "number of times to retry failed requests to vault")

	// upstream service type detection
	rootCmd.PersistentFlags().BoolVar(&discoveryOpts.AutoDiscoverSwagger, "detect-swagger-upstreams", true, "enable automatic discovery of upstreams that implement Swagger by querying for common Swagger Doc endpoints.")
	rootCmd.PersistentFlags().BoolVar(&discoveryOpts.AutoDiscoverNATS, "detect-nats-upstreams", true, "enable automatic discovery of upstreams that are running NATS by connecting to the default cluster id.")
	rootCmd.PersistentFlags().BoolVar(&discoveryOpts.AutoDiscoverGRPC, "detect-grpc-upstreams", true, "enable automatic discovery of upstreams that are running gRPC Services and haeve reflection enabled.")
	rootCmd.PersistentFlags().BoolVar(&discoveryOpts.AutoDiscoverFAAS, "detect-faas-upstreams", true, "enable automatic discovery open faas upstreams.")
	rootCmd.PersistentFlags().StringSliceVar(&discoveryOpts.SwaggerUrisToTry, "swagger-uris", []string{}, "paths function discovery should try to use to discover swagger services. function discovery will query http://<upstream>/<uri> for the swagger.json document. "+
		"if found, REST functions will be discovered for this upstream.")
}
