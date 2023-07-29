// Parse and filter Conda metadata JSON files
// https://github.com/conda/schemas
package main

import (
	"bytes"
	"flag"
	"log"
	"os"
	"path/filepath"
	"sort"

	"github.com/manics/go-conda-proxy/repodata"
)

func writeSortedSet(outputFilename string, s *repodata.Set) {
	log.Println("Writing", outputFilename)
	output, err := os.Create(outputFilename)
	if err != nil {
		log.Fatalf("Error opening file: %s", err)
	}
	defer output.Close()

	items := *s.Items()
	sort.Strings(items)
	for _, item := range items {
		if _, err := output.WriteString(item + "\n"); err != nil {
			log.Fatalf("Error writing to file: %s", err)
		}
	}
	log.Println("Output written to", outputFilename)
}

func main() {
	allFileNames := repodata.NewSet(nil)
	allPackageNames := repodata.NewSet(nil)

	configFile := flag.String("cfg", "", "Configuration file")
	forceUpdate := flag.Bool("force", false, "Force update")
	flag.Parse()

	if *configFile == "" {
		log.Fatalf("Configuration file required")
	}

	cfg, err := repodata.LoadCondaRepoConfig(*configFile)
	if err != nil {
		log.Fatalf("Failed to load configuration file: %s", err)
	}

	err = repodata.UpdateFromConfig(cfg, *forceUpdate)
	if err != nil {
		log.Fatalf("Failed to update repodata: %s", err)
	}

	outputPrefix := cfg.FilteredRepodataDir

	var allowedPackages *repodata.Set = nil
	var filteredRepodata map[string]*repodata.Repodata = make(map[string]*repodata.Repodata)

	for channel, channelCfg := range cfg.Channels {
		log.Printf("channel:[%s] channelCfg:[%+v]", channel, channelCfg)
		if channelCfg.AllowlistFile != "" {
			allowedPackages = repodata.ParseListFromFile(channelCfg.AllowlistFile)
			log.Printf("allowedPackages:[%d]", allowedPackages.Len())
		}

		for _, subdir := range channelCfg.Subdirs {
			file := repodata.GetDestinationFilename(cfg.OriginalRepodataDir, channel, subdir)
			filtered, fileNames, packageNames, err := repodata.ParseRepodata(channel, file, allowedPackages)
			if err != nil {
				log.Fatalf("Error parsing repodata: %s", err)
			}

			filteredRepodata[filtered.Info.Subdir] = filtered
			for _, k := range *fileNames.Items() {
				allFileNames.Add(k)
			}
			for _, k := range *packageNames.Items() {
				allPackageNames.Add(k)
			}

			filteredFile := repodata.GetDestinationFilename(outputPrefix, channel, subdir)
			data, err := repodata.EncodeJSON(filtered, " ")
			if err != nil {
				log.Fatalf("Error encoding JSON: %s", err)
			}

			err = repodata.WriteTempAndRename(bytes.NewReader(data), filteredFile)
			if err != nil {
				log.Fatalf("Error writing file: %s", err)
			}
		}
	}
	log.Printf("fileNames:[%d] packageNames:[%d]", allFileNames.Len(), allPackageNames.Len())

	// TODO: use WriteTempAndRename
	writeSortedSet(filepath.Join(outputPrefix, "filenames.txt"), allFileNames)
	writeSortedSet(filepath.Join(outputPrefix, "packagenames.txt"), allPackageNames)
}
