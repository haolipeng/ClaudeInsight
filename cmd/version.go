package cmd

import (
	"fmt"
	"claudeinsight/version"

	"github.com/spf13/cobra"
)

var (
	versionCmd = &cobra.Command{
		Use:   "version",
		Short: "Print the version of claudeinsight",
		Run: func(cmd *cobra.Command, args []string) {
			// The following vars are set by the linker during build. See the .goreleaser.yaml
			// reference: https://goreleaser.com/customization/builds/
			cmd.Println(fmt.Sprintf("Version: %v", version.GetVersion()))
			cmd.Println(fmt.Sprintf("BuildTime: %v", version.GetBuildTime()))
			cmd.Println(fmt.Sprintf("CommitID: %v", version.GetCommitID()))
		},
	}
)

func init() {
	rootCmd.AddCommand(versionCmd)
}
