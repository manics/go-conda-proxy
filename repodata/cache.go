package repodata

import (
	"errors"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

// WriteTempAndRename writes src to a tempfile, and then rename tempfile to destination
func WriteTempAndRename(src io.Reader, destination string) error {
	dir, fileout := filepath.Split(destination)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	temp, err := os.CreateTemp(dir, fileout+".tmp*")
	if err != nil {
		return err
	}
	defer temp.Close()

	if _, err := io.Copy(temp, src); err != nil {
		return err
	}

	err = os.Rename(temp.Name(), destination)
	return err
}

// UpdateDownload downloads a URL if it is older than maxAgeMinutes
//
// Set maxAgeMinutes to 0 to force an update
func UpdateDownload(url string, destination string, maxAgeMinutes int) error {
	if info, err := os.Stat(destination); err == nil {
		ageInMinutes := int(time.Since(info.ModTime()).Minutes())
		if maxAgeMinutes > 0 && ageInMinutes < maxAgeMinutes {
			log.Printf("Using cached %s (%d minutes old)\n", destination, ageInMinutes)
			return nil
		}
	}

	log.Printf("Updating %s from %s\n", destination, url)

	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	if resp.StatusCode != http.StatusOK {
		return errors.New(resp.Status + " " + url)
	}
	defer resp.Body.Close()

	err = WriteTempAndRename(resp.Body, destination)
	return err
}

func GetDestinationFilename(parentdir string, channel string, subdir string) string {
	return filepath.Join(parentdir, channel, subdir, "repodata.json")
}

func UpdateChannelRepodata(host string, parentdir string, channel string, subdirs []string, maxAgeMinutes int) error {
	errs := []error{}
	for _, subdir := range subdirs {
		destination := GetDestinationFilename(parentdir, channel, subdir)
		url := host + "/" + channel + "/" + subdir + "/repodata.json"
		if err := UpdateDownload(url, destination, maxAgeMinutes); err != nil {
			errs = append(errs, err)
			log.Printf("Error updating %s: %s\n", destination, err)
		}
	}
	return errors.Join(errs...)
}

func UpdateFromConfig(cfg *CondaRepoConfig, forceUpdate bool) error {
	errs := []error{}

	maxAgeMinutes := cfg.MaxAgeMinutes
	if forceUpdate {
		maxAgeMinutes = 0
	}
	for channel, channelConfig := range cfg.Channels {
		err := UpdateChannelRepodata(cfg.CondaHost, cfg.OriginalRepodataDir, channel, channelConfig.Subdirs, maxAgeMinutes)
		if err != nil {
			errs = append(errs, err)
		}
	}
	// Returns nil if all errs are nil
	return errors.Join(errs...)
}
