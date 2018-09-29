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
	"os/exec"
	"path/filepath"
	"time"

	"github.com/spf13/cobra"
)

// upCmd represents the up command
var upCmd = &cobra.Command{
	Use:   "up",
	Short: "Spin up a local bosh-lit VM with accessible BOSH director",
	Run: func(cmd *cobra.Command, args []string) {
		err := performUp()
		expectNoError(err)
	},
}

func init() {
	rootCmd.AddCommand(upCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// upCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// upCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

func performUp() error {
	command := exec.Command(
		"linuxkit", "run", "hyperkit",
		"-console-file",
		"-iso", "-uefi",
		"-cpus=4", "-mem=4096",
		"-disk", "size=80G",
		"-networking", "vpnkit",
		"-publish", "9999:9999/tcp",
		"-publish", "9998:9998/tcp",
		"-state", filepath.Join(bltHomeDir, "state", "linuxkit"),
		filepath.Join(bltHomeDir, "assets", "bosh-lit-efi.iso"),
	)

	err := command.Start()
	if err != nil {
		return err
	}

	err = vm.WaitForStatus(vm.VMStatusRunning, bltHomeDir, time.Minute)
	if err != nil {
		return err
	}

	fmt.Println("Success!")
	return nil
}