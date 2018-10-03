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
	"bufio"
	"fmt"
	"github.com/fatih/color"
	"github.com/mitchellh/go-homedir"
	"github.com/spf13/cobra"
	"os"
	"path/filepath"
	"strings"
)

var (
	bltHomeDir string
	boldWhite  = color.New(color.FgWhite, color.Bold)
	boldGreen  = color.New(color.FgGreen, color.Bold)
	boldYellow = color.New(color.FgYellow, color.Bold)
	boldRed    = color.New(color.FgRed, color.Bold)
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "blt",
	Short: "CLI for managing BOSH Lit VMs",
	Long: fmt.Sprintf(`
blt (short for BOSH Lit) is a CLI for managing a BOSH Lit VM.
This tool provides an easy way to get up and running with a local, low-footprint
BOSH environment that supports deployments.

%s

Website: https://github.com/aemengo/blt`, gettingStartedInstructions()),
	// Uncomment the following line if your bare application
	// has an action associated with it:
	//	Run: func(cmd *cobra.Command, args []string) { },
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	expectNoError(err)
}

func init() {
	cobra.OnInitialize(initConfig)

	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.
	//rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.blt.yaml)")

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	//rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

func initConfig() {
	home, err := homedir.Dir()
	expectNoError(err)

	if envHome := os.Getenv("BLT_HOME"); envHome != "" {
		home = envHome
	}

	bltHomeDir = filepath.Join(home, ".blt")
	err = os.MkdirAll(bltHomeDir, os.ModePerm)
	expectNoError(err)
}

func expectNoError(err error) {
	if err == nil {
		return
	}

	fmt.Printf(boldRed.Sprint("Error")+"\n%s.\n", err)
	os.Exit(1)
}

// confirmationCode adapted from:
// https://gist.github.com/r0l1/3dcbb0c8f6cfe9c66ab8008f55f8f28b
func askForConfirmation(s string, attempts int) bool {
	reader := bufio.NewReader(os.Stdin)

	for ; attempts > 0; attempts-- {
		fmt.Printf("%s %s: ", s, boldWhite.Sprintf(`[y/n]`))

		response, _ := reader.ReadString('\n')
		response = strings.TrimSpace(response)
		response = strings.ToLower(response)

		if response == "y" || response == "yes" {
			return true
		} else if response == "n" || response == "no" {
			return false
		}
	}

	return false
}

func exists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func gettingStartedInstructions() string {
	return fmt.Sprintf(`Getting Started
===============

Your BOSH director will be accessible at "10.0.0.4". To make sure your requests
target appropriately you may need to add the IP to your network interface, like so:

$ %s

Seeing your BOSH target credentials is as simple as using the "env" command. Thus in your
terminal, you can target like so:

$ %s

Networking
==========

BOSH Lit is provisioned with the network of "10.0.0.0/16". It also has a minimal cloud-config
pre-configured with the appropriate attributes. Should you wish to deploy Cloud Foundry
on your local instance, you can use this cloud-config here:

https://github.com/aemengo/bosh-runc-cpi-release/blob/master/operations/cf/cloud-config.yml

In order to access ports for new deployments from the host, you must use the "blt expose" command.
Port forwarding is done à la ssh. For more information, you can invoke the following:

$ %s

State
=====

The "state.json" and "creds.yml" of your BOSH director is by default held at: "$HOME/.blt/state/bosh".
For adequate state preservation between VM "ups" and "downs", direct access is NOT recommended.

To override where state files are kept, you may use the following environment variable:

$ %s

To see the message again, you can always run the blt CLI tool with no arguments:

$ %s`,
		boldWhite.Sprint("sudo ifconfig lo0 alias 10.0.0.4"),
		boldWhite.Sprintf(`eval "$(blt env)"`),
		boldWhite.Sprintf("blt expose -h"),
		boldWhite.Sprintf("export BLT_HOME=/path/to/dir"),
		boldWhite.Sprintf("blt"),
	)
}
