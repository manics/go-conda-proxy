// Parse and filter Conda metadata JSON files
// https://github.com/conda/schemas
package repodata

import (
	"bufio"
	"encoding/json"
	"io"
	"log"
	"os"
	"strings"
)

// packageIsAllowed returns true if the name is in allowedPackages, or if allowedPackages is nil
func packageIsAllowed(name string, allowedPackages *Set) bool {
	if allowedPackages == nil {
		return true
	}
	return allowedPackages.Contains(name)
}

// FilenameIsValid returns true if the filename matches the associated metadata
//
// > Filename key of each package should be validated against {name}-{version}-{build}{ext} metadata for the package
// https://github.com/conda/schemas/blob/bd2b05d6a6314b39d9a8c9c9802280c3eb78e788/repodata-1.schema.json#L33
func FilenameIsValid(filename string, expectedExt string, repodata *Repodata, record *RepodataRecord) bool {
	if record.Subdir != repodata.Info.Subdir {
		log.Printf("Subdir mismatch [%s]: %s != %s", filename, record.Subdir, repodata.Info.Subdir)
		return false
	}

	expectedFilename := record.Name + "-" + record.Version + "-" + record.Build + expectedExt
	if filename != expectedFilename {
		log.Printf("Filename does not match metadata: %s != %s", filename, expectedFilename)
		return false
	}
	return true
}

// ParseRepodata parses a Conda repodata JSON file, and filters it by allowedPackages
func ParseRepodata(channel string, repodataFile string, allowedPackages *Set) (*Repodata, *Set, *Set, error) {
	log.Println("Parsing", repodataFile)
	jsonFile, err := os.Open(repodataFile)
	if err != nil {
		log.Printf("Error opening file: %s", err)
		return nil, nil, nil, err
	}

	defer jsonFile.Close()

	var repodata Repodata
	byteValue, _ := io.ReadAll(jsonFile)
	if err := json.Unmarshal(byteValue, &repodata); err != nil {
		log.Printf("Error parsing JSON: %s", err)
		return nil, nil, nil, err
	}

	log.Printf("%s packages:[%d] packages.conda:[%d]", repodataFile, len(repodata.Packages), len(repodata.PackagesConda))

	filtered := FilterRepodataByAllowed(&repodata, allowedPackages)

	fileNames := NewSet(nil)
	packageNames := NewSet(nil)

	for _, records := range []map[string]RepodataRecord{filtered.Packages, filtered.PackagesConda} {
		for k, v := range records {
			fileNames.Add(channel + "/" + v.Subdir + "/" + k)
			packageNames.Add(v.Name)
		}
	}

	log.Printf("fileNames:[%d] packageNames:[%d]", fileNames.Len(), packageNames.Len())
	return filtered, fileNames, packageNames, nil
}

// FilterRepodataByAllowed checks the repodata and filters packages by allowed
func FilterRepodataByAllowed(repodata *Repodata, allowed *Set) *Repodata {
	// Shallow copy, apart from Packages and PackagesConda
	filtered := Repodata{
		RepodataVersion: repodata.RepodataVersion,
		Info:            repodata.Info,
	}
	filtered.Packages = make(map[string]RepodataRecord)
	filtered.PackagesConda = make(map[string]RepodataRecord)

	fileNames := NewSet(nil)
	packageNames := NewSet(nil)

	for k, v := range repodata.Packages {
		if !FilenameIsValid(k, ".tar.bz2", &filtered, &v) {
			log.Println("Filename is not valid", k)
			continue
		}
		if packageIsAllowed(v.Name, allowed) {
			fileNames.Add(v.Subdir + "/" + k)
			packageNames.Add(v.Name)
			filtered.Packages[k] = v
		}
	}

	for k, v := range repodata.PackagesConda {
		if !FilenameIsValid(k, ".conda", &filtered, &v) {
			log.Println("Filename is not valid", k)
			continue
		}
		if packageIsAllowed(v.Name, allowed) {
			fileNames.Add(v.Subdir + "/" + k)
			packageNames.Add(v.Name)
			filtered.PackagesConda[k] = v
		}
	}

	return &filtered
}

// ParseListFromFile parses a plain text file with a list of strings
//
// File should contain one string per line.
// Leading/trailing whitespace is stripped.
// Lines starting with '#' are ignored.
func ParseListFromFile(allowedFile string) *Set {
	allowedTxt, err := os.Open(allowedFile)
	if err != nil {
		log.Fatalf("Error opening file: %s", err)
	}
	defer allowedTxt.Close()

	allowedPackages := NewSet(nil)
	scanner := bufio.NewScanner(allowedTxt)
	for scanner.Scan() {
		name := strings.TrimSpace(scanner.Text())
		if len(name) > 0 && name[0] != '#' {
			allowedPackages.Add(name)
		}
	}

	return allowedPackages
}
