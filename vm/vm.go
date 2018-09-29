package vm

import (
	"context"
	"fmt"
	c1 "github.com/aemengo/bosh-runc-cpi/client"
	c2 "github.com/aemengo/vpnkit-manager/client"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"syscall"
	"time"
)

const (
	VMStatusRunning Status = iota
	VMStatusStopped
	VMStatusUnresponsive
)

type Status int

func (s Status) String() string {
	switch s {
	case VMStatusRunning:
		return "Running"
	case VMStatusStopped:
		return "Stopped"
	case VMStatusUnresponsive:
		return "Unresponsive"
	}

	return "Unknown"
}

func GetStatus(homedir string) Status {
	_, ok := fetchVMProcess(homedir)
	if !ok {
		return VMStatusStopped
	}

	ctx := context.Background()

	err := c1.Ping(ctx, "127.0.0.1:9999")
	if err != nil {
		return VMStatusUnresponsive
	}

	err = c2.Ping(ctx, "127.0.0.1:9998")
	if err != nil {
		return VMStatusUnresponsive
	}

	return VMStatusRunning
}

func WaitForStatus(desiredStatus Status, homedir string, timeout time.Duration) error {
	for {
		select {
		case <-time.After(timeout):
			return fmt.Errorf("vm failed to reach a status of %s after %v", desiredStatus, timeout)
		case <-time.NewTicker(time.Second).C:
			status := GetStatus(homedir)
			if status == desiredStatus {
				return nil
			}
		}
	}
}

func Stop(homedir string) {
	process, ok := fetchVMProcess(homedir)
	if !ok {
		return
	}

	process.Signal(os.Interrupt)

	err := WaitForStatus(VMStatusStopped, homedir, 20*time.Second)
	if err == nil {
		return
	}

	fmt.Println("VM did not terminate gracefully after 20 seconds. Force quitting...")
	process.Signal(os.Kill)
}

func fetchVMProcess(homedir string) (*os.Process, bool) {
	pidFile := filepath.Join(homedir, "state", "linuxkit", "hyperkit.pid")

	_, err := os.Stat(filepath.Join(homedir, "state", "linuxkit", "hyperkit.pid"))
	if os.IsNotExist(err) {
		return nil, false
	}

	if err != nil {
		return nil, false
	}

	data, err := ioutil.ReadFile(pidFile)
	if err != nil {
		return nil, false
	}

	pid, err := strconv.Atoi(string(data))
	if err != nil {
		return nil, false
	}

	process, err := os.FindProcess(pid)
	if err != nil {
		return nil, false
	}

	err = process.Signal(syscall.Signal(0))
	if err != nil {
		return nil, false
	}

	return process, true
}
