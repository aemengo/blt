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
	"os"
	"path/filepath"
)

// destroyCmd represents the destroy command
var destroyCmd = &cobra.Command{
	Use:   "destroy",
	Short: "Remove the saved state of your local bosh-lit VM",
	Run: func(cmd *cobra.Command, args []string) {
		err := performDestroy()
		expectNoError(err)
	},
}

func init() {
	rootCmd.AddCommand(destroyCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// destroyCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// destroyCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

func performDestroy() error {
	status := vm.GetStatus(bltHomeDir)
	if status != vm.VMStatusStopped {
		return fmt.Errorf("your VM must be stopped before you can perform this action, it is currently: %s", status)
	}

	ok := askForConfirmation("Do you really want to wipe all the data off of your bosh-lit VM?", 3)
	if !ok {
		fmt.Println("Aborting...")
		return nil
	}

	return os.RemoveAll(filepath.Join(bltHomeDir, "state", "linuxkit"))
}