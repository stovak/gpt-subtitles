/*
Copyright © 2025 TOM STOVALL <stovak @ gmail dot com>
*/
package drop

import (
	"fmt"

	"github.com/spf13/cobra"
)

// ListCmd represents the db/list command
var ListCmd = &cobra.Command{
	Use:   "drop:list",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("db:list called")
	},
}
