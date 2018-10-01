package web

import (
	"crypto/sha1"
	"fmt"
	"github.com/aemengo/blt/path"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"time"
)

func DownloadAssets(version string, homedir string) error {
	shaPath, err := fetch(path.AssetSHAurl(version))
	if err != nil {
		return err
	}
	defer os.RemoveAll(shaPath)

	data, err := ioutil.ReadFile(shaPath)
	if err != nil {
		return err
	}

	assetPath, err := fetch(path.AssetURL(version), string(data))
	if err != nil {
		return err
	}
	defer os.RemoveAll(assetPath)

	err = os.RemoveAll(path.AssetDir(homedir))
	if err != nil {
		return err
	}

	return unpackAsset(assetPath, homedir)
}

func unpackAsset(src string, homedir string) error {
	output, err := exec.Command("tar", "xf", src, "-C", homedir).CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to unpack assets to %s: %s: %s", homedir, err, output)
	}

	return nil
}

func fetch(url string, args ...string) (string, error) {
	var path string
	return path, do(5, 5*time.Second, func() error {
		tmpFile, err := ioutil.TempFile("", "blt-")
		if err != nil {
			return err
		}
		defer tmpFile.Close()

		path = tmpFile.Name()

		resp, err := http.Get(url)
		if err != nil {
			return err
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			return fmt.Errorf("received [%s] for %s", resp.Status, url)
		}

		_, err = io.Copy(tmpFile, resp.Body)
		if err != nil {
			return err
		}

		if len(args) > 0 {
			return shaMatches(args[0], path)
		} else {
			return nil
		}
	})
}

func shaMatches(sha string, path string) error {
	file, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("failed to open %s: %s", path, err)
	}
	defer file.Close()

	hash := sha1.New()
	_, err = io.Copy(hash, file)
	if err != nil {
		return fmt.Errorf("failed to read %s: %s", path, err)
	}

	actualSha := fmt.Sprintf("%x", hash.Sum([]byte{}))
	if actualSha != strings.TrimSpace(sha) {
		return fmt.Errorf("checksum validation error with: %s: ", path)
	}

	return nil
}

func do(attempts int, delay time.Duration, task func() error) (err error) {
	ticker := time.NewTicker(delay)

	for {
		select {
		case <-ticker.C:
			err = task()
			if err == nil {
				return
			}

			attempts--
			if attempts == 0 {
				return
			}
		}
	}
}