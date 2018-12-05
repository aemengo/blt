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
	"errors"
	"fmt"
	"github.com/aemengo/blt/path"
	"github.com/aemengo/blt/vm"
	"github.com/aemengo/blt/web"
	"io/ioutil"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

// upCmd represents the up command
var upCmd = &cobra.Command{
	Use:   "up",
	Short: "Spin up a local BOSH Lit VM with accessible BOSH director",
	Run: func(cmd *cobra.Command, args []string) {
		err := performUp()
		stopIndeterminateProgressAnimation()
		expectNoError(err)
	},
}

var (
	cpu      string
	memory   string
	disk     string
	doneChan = make(chan bool, 1)
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
	upCmd.Flags().StringVarP(&disk, "disk", "d", "40", "Amount of disk space to allocate to VM in gigabytes")
}

func performUp() error {
	status := vm.GetStatus(bltHomeDir)
	if status != vm.VMStatusStopped {
		fmt.Println("BOSH Lit is already running...")
		return nil
	}

	boldWhite.Print("Validating Prerequisites...   ")
	err := checkForDependencies()
	if err != nil {
		return err
	}

	err = checkNetworkAddrs()
	if err != nil {
		return err
	}
	boldGreen.Println("Success")

	startTime := time.Now()
	boldWhite.Print("Checking Assets...   ")

	ok := checkNeedsUpdates()
	if !ok {
		boldGreen.Println("Success")
	} else {
		boldYellow.Println("Needs Updates")

		messageChan := make(chan string, 10)
		go printMessages(messageChan)
		err = web.DownloadAssets(version, bltHomeDir, messageChan)
		if err != nil {
			return err
		}

		stopIndeterminateProgressAnimation()
	}

	err = os.RemoveAll(path.Pidpath(bltHomeDir))
	if err != nil {
		return err
	}

	boldWhite.Print("Starting VM")
	go showIndeterminateProgressAnimation()
	command := exec.Command(
		"linuxkit", "run", "hyperkit",
		"-console-file",
		"-iso", "-uefi",
		"-cpus="+cpu, "-mem="+memory,
		"-disk", "size="+disk+"G",
		"-networking", "vpnkit",
		"-vpnkit", filepath.Join(path.AssetDir(bltHomeDir), "vpnkit"),
		"-publish", "9999:9999/tcp",
		"-publish", "9998:9998/tcp",
		"-state", path.LinuxkitStatePath(bltHomeDir),
		path.EFIisoPath(bltHomeDir))

	logFile, err := os.Create(filepath.Join(bltHomeDir, "linuxkit.log"))
	if err != nil {
		return err
	}
	defer logFile.Close()

	command.Stdout = logFile
	command.Stderr = logFile

	err = command.Start()
	if err != nil {
		return err
	}

	err = vm.WaitForStatus(vm.VMStatusRunning, bltHomeDir, time.Minute)
	if err != nil {
		return err
	}

	stopIndeterminateProgressAnimation()
	boldGreen.Println("Success")

	err = resetBOSHStateJSON()
	if err != nil {
		return err
	}

	boldWhite.Println("Deploying Director...  ")
	command = exec.Command(
		"bosh", "create-env", filepath.Join(path.BoshDeploymentDir(bltHomeDir), "bosh.yml"),
		"-o", filepath.Join(path.BoshDeploymentDir(bltHomeDir), "jumpbox-user.yml"),
		"-o", filepath.Join(path.BoshOperationsDir(bltHomeDir), "runc-cpi.yml"),
		"--state", path.BoshStateJSONPath(bltHomeDir),
		"--vars-store", path.BoshCredsPath(bltHomeDir),
		"-v", "director_name=director",
		"-v", "external_cpid_ip=127.0.0.1",
		"-v", "internal_cpid_ip=192.168.65.3",
		"-v", "internal_cpid_gw=192.168.65.1",
		"-v", "internal_ip=10.0.0.4",
		"-v", "internal_gw=10.0.0.1",
		"-v", "internal_cidr=10.0.0.0/16")
	command.Stdout = os.Stdout
	command.Stderr = os.Stderr

	err = command.Run()
	if err != nil {
		return err
	}

	boldWhite.Printf("Configuring Director...  ")
	configureBoshDirector()
	boldGreen.Println("Success")

	boldGreen.Printf("\nCompleted in %v\n\n", time.Since(startTime))

	handleUsageMessage()
	return nil
}

func resetBOSHStateJSON() error {
	_, err := os.Stat(path.BoshStateJSONPath(bltHomeDir))
	if os.IsNotExist(err) {
		return nil
	}

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

func printMessages(messageChan chan string) {
	for {
		select {
		case <-doneChan:
			return
		case m := <-messageChan:
			fmt.Printf(m)
		}
	}
}
func showIndeterminateProgressAnimation() {
	var (
		toggle     bool
		clearChars = "\b\b\b\b\b\b"
		ticker     = time.NewTicker(500*time.Millisecond)
	)

	boldWhite.Print("...   ")

	for {
		select {
		case <-doneChan:
			return
		case <-ticker.C:
			if toggle {
				boldWhite.Print(clearChars, "...   ")
			} else {
				boldWhite.Print(clearChars, "..    ")
			}

			toggle = !toggle
		}
	}
}

func stopIndeterminateProgressAnimation() {
	doneChan <- true
}

func handleUsageMessage() {
	if exists(path.FirstBootMarker(bltHomeDir)) {
		return
	}

	fmt.Println(gettingStartedInstructions(), "\n")
	f, _ := os.Create(path.FirstBootMarker(bltHomeDir))
	if f != nil {
		f.Close()
	}
}

func configureBoshDirector() error {
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

	cmd := fmt.Sprintf(`eval "%s"; bosh cloud-config`, fetchEnvironmentVariables())
	err := exec.Command("bash", "-c", cmd).Run()
	if err == nil {
		return nil
	}

	cmd = fmt.Sprintf(`eval "%s"; bosh -n update-cloud-config %s`, fetchEnvironmentVariables(), filepath.Join(path.BoshOperationsDir(bltHomeDir), "cloud-config.yml"))
	output, err := exec.Command("bash", "-c", cmd).CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to execute '%s': %s: %s", cmd, err, output)
	}

	return nil
}

func checkNeedsUpdates() bool {
	if version == "DEV" {
		return false
	}

	contents, err := ioutil.ReadFile(path.AssetVersionPath(bltHomeDir))
	if err != nil {
		return true
	}

	return strings.TrimSpace(string(contents)) != version
}

func checkNetworkAddrs() error {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return fmt.Errorf("failed to inspect network interfaces: %s", err)
	}

	for _, addr := range addrs {
		elements := strings.Split(addr.String(), "/")
		if elements[0] == "10.0.0.4" {
			return nil
		}
	}

	return fmt.Errorf(`Your BOSH director will be accessible at %s. To make sure your requests
target appropriately you must add the IP to your network interfaces, like so:

$ %s

`, boldWhite.Sprint("10.0.0.4"), boldWhite.Sprint("sudo ifconfig lo0 alias 10.0.0.4"))
}

type Dependency struct {
	Name         string
	CheckCommand string
	Site         string
}

func (d *Dependency) Usage() string {
	if d.Site == "" {
		return boldWhite.Sprintf(strings.Title(d.Name))
	}

	return fmt.Sprintf("%s %s", boldWhite.Sprintf(strings.Title(d.Name)), d.Site)
}

func checkForDependencies() error {
	var missingDeps []Dependency
	var allDeps = []Dependency{
		{
			Name:         "docker",
			CheckCommand: "docker -v",
			Site:         "https://store.docker.com/editions/community/docker-ce-desktop-mac",
		},
		{
			Name:         "bosh",
			CheckCommand: "bosh -v",
			Site:         "https://bosh.io/docs/cli-v2",
		},
		{
			Name:         "linuxkit",
			CheckCommand: "linuxkit version",
			Site:         "https://github.com/linuxkit/linuxkit",
		},
		{
			Name:         "tar",
			CheckCommand: "tar --help",
		},
	}

	for _, d := range allDeps {
		if err := exec.Command("/bin/sh", "-c", d.CheckCommand).Run(); err != nil {
			missingDeps = append(missingDeps, d)
		}
	}

	if len(missingDeps) == 0 {
		return nil
	}

	var messages = []string{"The following dependencies must be installed:"}
	for i, d := range missingDeps {
		messages = append(messages, fmt.Sprintf("%d: %s", i, d.Usage()))
	}
	return errors.New(strings.Join(messages, "\n"))
}
