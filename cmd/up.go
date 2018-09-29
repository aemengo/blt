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
	"encoding/json"
	"fmt"
	"github.com/aemengo/blt/path"
	"github.com/aemengo/blt/vm"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/spf13/cobra"
)

// upCmd represents the up command
var upCmd = &cobra.Command{
	Use:   "up",
	Short: "Spin up a local BOSH Lit VM with accessible BOSH director",
	Run: func(cmd *cobra.Command, args []string) {
		err := performUp()
		expectNoError(err)
	},
}

var (
	cpu    string
	memory string
	disk   string
)

func init() {
	rootCmd.AddCommand(upCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// upCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// upCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")

	upCmd.Flags().StringVarP(&cpu, "cpu", "c", "4", "Number of cores to allocate to VM")
	upCmd.Flags().StringVarP(&memory, "memory", "m", "4096", "Amount of memory to allocate to VM in megabytes")
	upCmd.Flags().StringVarP(&disk, "disk", "d", "40", "Amount of disk space to allocate to VM in gigabytes (copy-on-write)")
}

func performUp() error {
	status := vm.GetStatus(bltHomeDir)
	if status != vm.VMStatusStopped {
		fmt.Println("BOSH Lit is already running...")
		return nil
	}

	os.RemoveAll(path.Pidpath(bltHomeDir))

	// validate dependencies
	// fetch assets

	// declare launching vm
	command := exec.Command(
		"linuxkit", "run", "hyperkit",
		"-console-file",
		"-iso", "-uefi",
		"-cpus="+cpu, "-mem="+memory,
		"-disk", "size="+disk+"G",
		"-networking", "vpnkit",
		"-publish", "9999:9999/tcp",
		"-publish", "9998:9998/tcp",
		"-state", path.LinuxkitStatePath(bltHomeDir),
		path.EFIisoPath(bltHomeDir),
	)

	err := command.Start()
	if err != nil {
		return err
	}

	err = vm.WaitForStatus(vm.VMStatusRunning, bltHomeDir, time.Minute)
	if err != nil {
		return err
	}

	err = resetBOSHStateJSON()
	if err != nil {
		return err
	}

	// declare deploying director
	command = exec.Command(
		"bosh", "create-env", filepath.Join(path.BoshDeploymentDir(bltHomeDir), "bosh.yml"),
		"-o", filepath.Join(path.BoshDeploymentDir(bltHomeDir), "jumpbox-user.yml"),
		"-o", filepath.Join(path.BoshOperationsDir(bltHomeDir), "runc-cpi.yml"),
		"--state", path.BoshStateJSONPath(bltHomeDir),
		"--vars-store", path.BoshCredsPath(bltHomeDir),
		"-v", "director_name=director",
		"-v", "external_cpid_ip=127.0.0.1",
		"-v", "internal_cpid_ip=192.168.65.3",
		"-v", "internal_ip=10.0.0.4",
		"-v", "internal_gw=10.0.0.1",
		"-v", "internal_cidr=10.0.0.0/16")
	command.Stdout = os.Stdout
	command.Stderr = os.Stderr

	err = command.Run()
	if err != nil {
		return err
	}

	commands := []string{
		fmt.Sprintf("bosh int %s --path /director_ssl/ca > %s", path.BoshCredsPath(bltHomeDir), path.BoshCACertPath(bltHomeDir)),
		fmt.Sprintf("bosh int %s --path /jumpbox_ssh/private_key > %s", path.BoshCredsPath(bltHomeDir), path.BoshGWPrivateKeyPath(bltHomeDir)),
		fmt.Sprintf("chmod 0600 %s", path.BoshGWPrivateKeyPath(bltHomeDir)),
	}

	for _, command := range commands {
		output, err := exec.Command("bash", "-c", command).CombinedOutput()
		if err != nil {
			return fmt.Errorf("failed to execute '%s': %s: %s", command, err, output)
		}
	}

	err = exec.Command("bash", "-c", `eval "%s"; bosh cloud-config`).Run()
	if err != nil {
		cmd := fmt.Sprintf(`eval "%s"; bosh -n update-cloud-config %s`, fetchEnvironmentVariables(), filepath.Join(path.BoshOperationsDir(bltHomeDir), "cloud-config.yml"))
		output, err := exec.Command("bash", "-c", cmd).CombinedOutput()
		if err != nil {
			return fmt.Errorf("failed to execute '%s': %s: %s", cmd, err, output)
		}
	}

	fmt.Println("We doing it!!!!")

	// add getting-started steps
	return nil
}

func resetBOSHStateJSON() error {
	mapping := map[string]interface{}{}

	data, err := ioutil.ReadFile(path.BoshStateJSONPath(bltHomeDir))
	if err != nil {
		return fmt.Errorf("failed to read bosh state file: %s", err)
	}

	json.Unmarshal(data, &mapping)
	delete(mapping, "current_manifest_sha")
	newContents, _ := json.Marshal(mapping)

	return ioutil.WriteFile(path.BoshStateJSONPath(bltHomeDir), newContents, 0600)
}