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
	"github.com/aemengo/blt/path"
	"github.com/spf13/cobra"
)

// envCmd represents the env command
var envCmd = &cobra.Command{
	Use:   "env",
	Short: "Output environment variables for accessing your BOSH director",
	Run: func(cmd *cobra.Command, args []string) {
		performEnv()
	},
}

func init() {
	rootCmd.AddCommand(envCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// envCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// envCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

func performEnv() {
	fmt.Printf(`export BOSH_ENVIRONMENT="10.0.0.4"
export BOSH_CLIENT=admin
export BOSH_CLIENT_SECRET="%s"
export BOSH_CA_CERT=%s
export BOSH_GW_HOST="10.0.0.4"
export BOSH_GW_USER="jumpbox"
export BOSH_GW_PRIVATE_KEY=%s
`,"mysecret", path.BoshCACertPath(bltHomeDir), path.BoshGWPrivateKeyPath(bltHomeDir))
}
