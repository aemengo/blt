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

// exposeCmd represents the expose command
var exposeCmd = &cobra.Command{
	Use:   "expose",
	Short: "Forward a port from the VM to the host",
	Long: `In order to access container ports from the host, they
must be explicitly forwarded. Declaring ports is done in
a similar fashion to the ssh command.

    -L host_address:port:container_address:container_port

For example:
$ blt expose -L 10.0.0.5:80:10.0.0.5:80 -L 10.0.0.5:443:10.0.0.5:443

Note: It is still up to you to configure how your machine routes itself to the
exposed port.
`,
	Run: func(cmd *cobra.Command, args []string) {
		err := performExpose()
		expectNoError(err)
	},
}

var addresses []string

func init() {
	rootCmd.AddCommand(exposeCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// exposeCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// exposeCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")

	exposeCmd.Flags().StringSliceVarP(&addresses, "forward", "L", []string{}, "List of addresses to forward from VM to host (required)")
	exposeCmd.MarkFlagRequired("forward")
}

func performExpose() error {
	status := vm.GetStatus(bltHomeDir)
	if status != vm.VMStatusRunning {
		return fmt.Errorf("your VM must be running before you can perform this action, it is currently: %s", boldWhite.Sprint(status))
	}

	return vm.Forward(addresses)
}