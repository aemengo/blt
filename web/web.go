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

func DownloadAssets(version string, homedir string, messageChan chan string) error {
	shaPath, err := fetch(path.AssetSHAurl(version), path.AssetSHAApproximateSize(), messageChan)
	if err != nil {
		return err
	}
	defer os.RemoveAll(shaPath)

	data, err := ioutil.ReadFile(shaPath)
	if err != nil {
		return err
	}

	assetPath, err := fetch(path.AssetURL(version), path.AssetsApproximateSize(), messageChan, string(data))
	if err != nil {
		return err
	}
	defer os.RemoveAll(assetPath)

	err = os.RemoveAll(path.AssetDir(homedir))
	if err != nil {
		return err
	}

	messageChan <- fmt.Sprintf("Unpacking assets into %s", homedir)
	return unpackAsset(assetPath, homedir)
}

func unpackAsset(src string, homedir string) error {
	output, err := exec.Command("tar", "xf", src, "-C", homedir).CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to unpack assets to %s: %s: %s", homedir, err, output)
	}

	return nil
}

func fetch(url string, size string, messageChan chan string, args ...string) (string, error) {
	var (
		path string
		retries = 4
	)

	return path, do(retries, 5*time.Second, func(attempt int) error {
		tmpFile, err := ioutil.TempFile("", "blt-")
		if err != nil {
			return err
		}
		defer tmpFile.Close()

		path = tmpFile.Name()

		messageChan <- fmt.Sprintf("[%d/%d][%s] Fetching %s", attempt, retries+1, size, url)
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
			messageChan <- fmt.Sprintf("[%d/%d] Performing integrity validation", attempt, retries+1)
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

func do(retries int, delay time.Duration, task func(int) error) (err error) {
	var counter = 1

	err = task(counter)
	if err == nil {
		return
	}

	ticker := time.NewTicker(delay)

	for {
		select {
		case <-ticker.C:
			counter++
			err = task(counter)
			if err == nil {
				return
			}

			retries--
			if retries == 0 {
				return
			}
		}
	}
}