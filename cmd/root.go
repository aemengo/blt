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
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "blt",
	Short: "CLI for managing BOSH Lit VMs",
	Long: `blt (short for BOSH Lit) is a CLI library for managing a BOSH Lit VM.
This tool provides an easy way to get up and running with a local, low-footprint
BOSH environment which supports deployments.

Website: https://github.com/aemengo/blt`,
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
}

func expectNoError(err error) {
	if err == nil {
		return
	}

	fmt.Printf(color.RedString("Error") + ": %s.\n", err)
	os.Exit(1)
}

// confirmationCode adapted from:
// https://gist.github.com/r0l1/3dcbb0c8f6cfe9c66ab8008f55f8f28b
func askForConfirmation(s string, attempts int) bool {
	reader := bufio.NewReader(os.Stdin)

	for ; attempts > 0; attempts-- {
		fmt.Printf("%s [y/n]: ", s)

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

func gettingStartedText() string {
	return fmt.Sprintf(`Getting Started
===============

Your bosh director


`)
}