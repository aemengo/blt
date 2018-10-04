package path

import (
	"fmt"
	"path/filepath"
)

func StateDir(homedir string) string {
	return filepath.Join(homedir, "state")
}

func AssetDir(homedir string) string {
	return filepath.Join(homedir, "assets")
}

func FirstBootMarker(homedir string) string {
	return filepath.Join(StateDir(homedir), "blt")
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
	return filepath.Join(AssetDir(homedir), "bosh-lit-efi.iso")
}

func BoshDeploymentDir(homedir string) string {
	return filepath.Join(AssetDir(homedir), "bosh-deployment")
}

func BoshOperationsDir(homedir string) string {
	return filepath.Join(AssetDir(homedir), "operations")
}

func AssetVersionPath(homedir string) string {
	return filepath.Join(AssetDir(homedir),  "version")
}

func AssetURL(version string) string {
	return fmt.Sprintf("https://github.com/aemengo/blt/releases/download/%s/bosh-lit-assets.tgz", version)
}

func AssetSHAurl(version string) string {
	return fmt.Sprintf("https://github.com/aemengo/blt/releases/download/%s/bosh-lit-assets.tgz.sha1", version)
}
