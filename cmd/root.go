package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "trix",
	Short: "Kubernetes security scanner",
	Long: `trix scans your Kubernetes clusters for vulnerabilities
and compliance issues using Trivy and custom CIS checks.`,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
