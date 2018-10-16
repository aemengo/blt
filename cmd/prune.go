// Copyright © 2018 NAME HERE <EMAIL ADDRESS>
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cmd

import (
	"fmt"
	"github.com/aemengo/blt/vm"
	"github.com/spf13/cobra"
)

// pruneCmd represents the prune command
var pruneCmd = &cobra.Command{
	Use:   "prune",
	Short: "Perform a discard operation on unused blocks in the VM",
	Long: `Perform a discard operation on unused blocks in the VM.

Typically useful for reclaiming disk space after deleting multiple BOSH deployments.`,
	Run: func(cmd *cobra.Command, args []string) {
		err := performPrune()
		expectNoError(err)
	},
}

func init() {
	rootCmd.AddCommand(pruneCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// pruneCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// pruneCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

func performPrune() error {
	status := vm.GetStatus(bltHomeDir)
	if status != vm.VMStatusRunning {
		return fmt.Errorf("your VM must be running before you can perform this action, it is currently: %s", boldWhite.Sprint(status))
	}

	return vm.Prune()
}