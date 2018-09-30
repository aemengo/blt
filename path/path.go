package path

import "path/filepath"

func StateDir(homedir string) string {
	return filepath.Join(homedir, "state")
}

func LinuxkitStatePath(homedir string) string {
	return filepath.Join(StateDir(homedir), "linuxkit")
}

func BoshStatePath(homedir string) string {
	return filepath.Join(StateDir(homedir), "bosh")
}

func BoshCredsPath(homedir string) string {
	return filepath.Join(BoshStatePath(homedir), "creds.yml")
}

func BoshStateJSONPath(homedir string) string {
	return filepath.Join(BoshStatePath(homedir), "state.json")
}

func BoshCACertPath(homedir string) string {
	return filepath.Join(BoshStatePath(homedir), "ca.crt")
}

func BoshGWPrivateKeyPath(homedir string) string {
	return filepath.Join(BoshStatePath(homedir), "gw_id_rsa")
}

func Pidpath(homedir string) string {
	return filepath.Join(LinuxkitStatePath(homedir), "hyperkit.pid")
}

func EFIisoPath(homedir string) string {
	return filepath.Join(homedir, "assets", "bosh-lit-efi.iso")
}

func BoshDeploymentDir(homedir string) string {
	return filepath.Join(homedir, "assets", "bosh-deployment")
}

func BoshOperationsDir(homedir string) string {
	return filepath.Join(homedir, "assets", "operations")
}