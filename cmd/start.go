package cmd

import (
	"fmt"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"k8s.io/client-go/rest"
)

func startCmd() *cobra.Command {
	var port int
	cmd := &cobra.Command{
		Use:   "start",
		Short: "Start Glue Function Discovery service",
		RunE: func(c *cobra.Command, args []string) error {
			cfg, err := getClientConfig()
			if err != nil {
				return errors.Wrap(err, "unable to get client configuration")
			}
			start(cfg, port)
			return nil
		},
	}
	cmd.LocalFlags().IntVarP(&port, "port", "p", 8080, "Port. If not set tries PORT environment variable before defaulting to 8080")
	return cmd
}

func start(cfg *rest.Config, port int) {
	// get list of sources for funcctions
	fmt.Println("starting server at port ", port)
}
