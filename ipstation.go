package stationip

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/yangchenxing/go-regionid"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"time"
)

type Config struct {
	LocalPath     string
	RemotePath    string
	CheckInterval time.Duration
	ErrorCallback func(error)
}

var (
	ipIndex           *index
	errNotModified    = errors.New("not modified")
	errNoFile         = errors.New("no local file and remote url")
	errNotInitialized = errors.New("not initialized")
)

func Initialize(config Config) error {
	if !regionid.Initialized() {
		if err := regionid.LoadBuildWorld(); err != nil {
			return fmt.Errorf("load buildin regionid fail: %s", err.Error())
		}
	}
	if err := guardLocal(config.LocalPath, config.RemotePath); err != nil {
		return fmt.Errorf("guard local file fail: %s", err.Error())
	}
	if err := load(config.LocalPath); err != nil {
		return fmt.Errorf("load local file fail: %s", err.Error())
	}
	if config.CheckInterval > 0 {
		if config.RemotePath != "" {
			go autoUpdateRemote(config.LocalPath, config.RemotePath, config.CheckInterval, config.ErrorCallback)
		} else {
			go autoUpdateLocal(config.LocalPath, config.CheckInterval, config.ErrorCallback)
		}
	}
	return nil
}

func Search(ip net.IP) (Result, error) {
	if ipIndex == nil {
		return notFound, errNotInitialized
	}
	return ipIndex.search(ip), nil
}

func MustSearch(ip net.IP) Result {
	result, _ := Search(ip)
	return result
}

func guardLocal(local, remote string) error {
	if _, err := os.Stat(local); err == nil {
		return nil
	}
	if remote == "" {
		return errNoFile
	}
	return download(remote, local, "")
}

func autoUpdateRemote(local, remote string, interval time.Duration, errorCallback func(error)) {
	for {
		if err := updateRemote(local, remote); err != nil {
			if errorCallback != nil {
				errorCallback(err)
			}
		}
		time.Sleep(interval)
	}
}

func autoUpdateLocal(local string, interval time.Duration, errorCallback func(error)) {
	info, err := os.Stat(local)
	if err != nil {
		errorCallback(err)
	}
	ts := info.ModTime()
	for {
		time.Sleep(interval)
		info, err := os.Stat(local)
		if err != nil {
			errorCallback(err)
			continue
		}
		if !info.ModTime().After(ts) {
			continue
		}
		if err = load(local); err != nil {
			errorCallback(err)
		} else {
			ts = info.ModTime()
		}
	}
}

func updateRemote(local, remote string) error {
	if err := download(remote, local, ipIndex.ETag); err == errNotModified {
		return nil
	} else if err != nil {
		return err
	}
	return load(local)
}

func load(local string) error {
	etag, err := ioutil.ReadFile(local + ".etag")
	if err != nil {
		return fmt.Errorf("read etag fail: %s", err.Error())
	}
	file, err := os.Open(local)
	if err != nil {
		return fmt.Errorf("open data fail: %s", err.Error())
	}
	defer file.Close()
	newIndex, err := newIndex(file, string(etag))
	if err != nil {
		return fmt.Errorf("load data fail: %s", err.Error())
	}
	ipIndex = newIndex
	return nil
}

func download(remote, local, etag string) error {
	resp, err := http.Get(remote)
	if err != nil {
		return fmt.Errorf("download fail: %s", err.Error())
	}
	defer resp.Body.Close()
	e := resp.Header.Get("ETag")
	if e != "" && etag == e {
		return errNotModified
	}
	if err := saveStreamToFile(resp.Body, local); err != nil {
		return fmt.Errorf("save data file fail: %s", err.Error())
	}
	if err := saveStreamToFile(bytes.NewBufferString(e), local+".etag"); err != nil {
		return fmt.Errorf("save etag file fail: %s", err.Error())
	}
	return nil
}

func saveStreamToFile(r io.Reader, local string) error {
	tempLocal := local + ".tmp"
	tempFile, err := os.OpenFile(tempLocal, os.O_WRONLY|os.O_CREATE, 0755)
	if err != nil {
		return err
	}
	if _, err := io.Copy(tempFile, r); err != nil {
		tempFile.Close()
		os.Remove(tempLocal)
		return err
	}
	tempFile.Close()
	return os.Rename(local, tempLocal)
}
