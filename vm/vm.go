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

	return ""
}

func GetStatus(homeDir string) Status {
	pidFile := filepath.Join(homeDir, "state", "linuxkit", "hyperkit.pid")
	if !pidExists(pidFile) {
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

func pidExists(pidFile string) bool {
	_, err := os.Stat(pidFile)
	if os.IsNotExist(err) {
		return false
	}

	if err != nil {
		return false
	}

	data, _ := ioutil.ReadFile(pidFile)
	pid, _ := strconv.Atoi(string(data))

	process, err := os.FindProcess(pid)
	if err != nil {
		return false
	}

	err = process.Signal(syscall.Signal(0))
	return err == nil
}
