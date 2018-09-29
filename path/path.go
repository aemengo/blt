package path

import "path/filepath"

func LinuxkitStatePath(homedir string) string {
	return filepath.Join(homedir, "state", "linuxkit")
}

func BoshStatePath(homedir string) string {
	return filepath.Join(homedir, "state", "bosh")
}

func Pidpath(homedir string) string {
	return filepath.Join(LinuxkitStatePath(homedir), "hyperkit.pid")
}

func EFIisoPath(homedir string) string {
	return filepath.Join(homedir, "assets", "bosh-lit-efi.iso")
}

func BoshCACertPath(homedir string) string {
	return filepath.Join(homedir, "state", "bosh", "ca.crt")
}

func BoshGWPrivateKeyPath(homedir string) string {
	return filepath.Join(homedir, "state", "bosh", "gw_id_rsa")
}

func BoshDeploymentDir(homedir string) string {
	return filepath.Join(homedir, "assets", "bosh-deployment")
}

func BoshOperationsDir(homedir string) string {
	return filepath.Join(homedir, "assets", "operations")
}