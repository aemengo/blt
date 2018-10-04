package web

import (
	"crypto/sha1"
	"fmt"
	"github.com/aemengo/blt/path"
	"github.com/dustin/go-humanize"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

func DownloadAssets(version string, homedir string, messageChan chan string) error {
	shaPath, err := fetch(path.AssetSHAurl(version), messageChan)
	if err != nil {
		return err
	}
	defer os.RemoveAll(shaPath)

	data, err := ioutil.ReadFile(shaPath)
	if err != nil {
		return err
	}

	assetPath, err := fetch(path.AssetURL(version), messageChan, string(data))
	if err != nil {
		return err
	}
	defer os.RemoveAll(assetPath)

	err = os.RemoveAll(path.AssetDir(homedir))
	if err != nil {
		return err
	}

	messageChan <- fmt.Sprintf("Unpacking assets into %s\n", homedir)
	return unpackAsset(assetPath, homedir)
}

func unpackAsset(src string, homedir string) error {
	output, err := exec.Command("tar", "xf", src, "-C", homedir).CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to unpack assets to %s: %s: %s", homedir, err, output)
	}

	return nil
}

func fetch(url string, messageChan chan string, args ...string) (string, error) {
	var (
		path     string
		retries  = 4
		doneChan = make(chan bool, 1)

		// Error'd Channel helps distinguish behavior from
		// when the Done Channel is invoked
		errdChan = make(chan bool, 1)

		// The Ready Channel lets us time
		// when to continue printing correctly
		readyChan = make(chan bool, 1)
	)

	return path, do(retries, 5*time.Second, func(attempt int) error {
		tmpFile, err := ioutil.TempFile("", "blt-")
		if err != nil {
			return err
		}
		defer tmpFile.Close()

		path = tmpFile.Name()

		startTime := time.Now()

		resp, err := http.Get(url)
		if err != nil {
			return err
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			return fmt.Errorf("received [%s] for %s", resp.Status, url)
		}

		size, _ := strconv.ParseUint(resp.Header.Get("Content-Length"), 10, 32)

		go func() {
			ticker := time.NewTicker(50 * time.Millisecond)

			for {
				select {
				case <-errdChan:
					messageChan <- "\n"
					readyChan <- true
					return
				case <-doneChan:
					messageChan <- fmt.Sprintf("\r\033[K[%d/%d] (%s/%s) Downloading '%s'... %v", attempt, retries+1, humanize.Bytes(size), humanize.Bytes(size), filepath.Base(url), time.Since(startTime))
					messageChan <- "\n"
					readyChan <- true
					return
				case <-ticker.C:
					fi, _ := tmpFile.Stat()
					messageChan <- fmt.Sprintf("\r\033[K[%d/%d] (%s/%s) Downloading '%s'... ", attempt, retries+1, humanize.Bytes(uint64(fi.Size())), humanize.Bytes(size), filepath.Base(url))
				}
			}
		}()

		_, err = io.Copy(tmpFile, resp.Body)
		if err != nil {
			errdChan <- true
			<-readyChan
			return err
		}

		doneChan <- true
		<-readyChan

		if len(args) > 0 {
			messageChan <- fmt.Sprintf("[%d/%d] Performing integrity validation\n", attempt, retries+1)
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
		return fmt.Errorf("checksum validation error with: %s", path)
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
