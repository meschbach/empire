package main

import (
	"context"
	"fmt"
	"github.com/spf13/cobra"
	"os"
)

func main() {
	service := &app{}
	root := &cobra.Command{
		Use: "empired",
		Run: func(cmd *cobra.Command, args []string) {
			procContext, done := context.WithCancel(cmd.Context())
			defer done()
			if err := service.Serve(procContext); err != nil {
				fmt.Printf("Failed to run: %e", err)
			}
		},
	}
	root.Flags().StringVarP(&service.base, "base", "b", "etc", "Configuration base to pickup")
	root.Flags().StringVarP(&service.fsRoot, "fs-root", "f", "", "Base to execute at")
	root.Flags().StringVarP(&service.boxed, "boxed", "j", "", "set the root of the output file system")

	if err := root.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
	}
}
