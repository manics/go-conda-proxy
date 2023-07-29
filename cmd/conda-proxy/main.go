// Based on https://gist.github.com/yowu/f7dc34bd4736a65ff28d
package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	"github.com/manics/go-conda-proxy/repodata"
	"golang.org/x/exp/slices"
)

const CONDA_HOST = "https://conda.anaconda.org"

var CONDA_CHANNELS = []string{"conda-forge"}
var CONDA_SUBDIRS = []string{"linux-64", "linux-aarch64", "osx-64", "osx-arm64", "win-64", "win-arm64", "noarch"}

const SHORT_TIMEOUT = 5 * time.Second
const LONG_TIMEOUT = 120 * time.Second

func copyHeader(dst, src http.Header) {
	for k, vv := range src {
		for _, v := range vv {
			dst.Add(k, v)
		}
	}
}

type proxy struct {
	AllowedFilenames *repodata.Set
	Cfg              *repodata.CondaRepoConfig
}

func httpLogPrefix(req *http.Request) string {
	return strings.Split(req.RemoteAddr, ":")[0]
}

func (p *proxy) serveRepodata(wr http.ResponseWriter, req *http.Request, channel string, subdir string, filename string) {
	logPrefix := httpLogPrefix(req)
	filePath := strings.Join([]string{channel, subdir, filename}, "/")

	// Ignore current_repodata.json, conda should fallback to repodata.json
	// https://docs.conda.io/projects/conda-build/en/stable/concepts/generating-index.html#trimming-to-current-repodata
	if filename != "repodata.json" {
		msg := "Invalid path: " + req.URL.Path
		http.Error(wr, msg, http.StatusNotFound)
		log.Println(logPrefix, http.StatusNotFound, msg)
		return
	}

	_, ok := p.Cfg.Channels[channel]
	if !ok || !slices.Contains(p.Cfg.Channels[channel].Subdirs, subdir) {
		msg := "Invalid channel/subdir: " + filePath
		http.Error(wr, msg, http.StatusNotFound)
		log.Println(logPrefix, http.StatusNotFound, msg)
		return
	}

	localPath := repodata.GetDestinationFilename(p.Cfg.FilteredRepodataDir, channel, subdir)

	wr.Header().Set("Content-Type", "application/json")
	wr.Header().Set("Cache-Control", fmt.Sprintf("max-age=%d", p.Cfg.CacheControlMaxAgeMinutes*60))

	http.ServeFile(wr, req, localPath)
}

func (p *proxy) ServeHTTP(wr http.ResponseWriter, req *http.Request) {
	logPrefix := httpLogPrefix(req)
	log.Println(logPrefix, req.Method, req.URL, req.UserAgent())

	if req.Method != "GET" {
		msg := "Invalid method: " + req.Method
		http.Error(wr, msg, http.StatusBadRequest)
		log.Println(logPrefix, http.StatusBadRequest, msg)
		return
	}

	pathParts := strings.Split(req.URL.Path, "/")
	filePath := strings.Join(pathParts[1:], "/")
	// first element should be empty due to the leading /
	if pathParts[0] != "" {
		msg := "Invalid filepath: " + req.URL.Path
		http.Error(wr, msg, http.StatusNotFound)
		log.Println(logPrefix, http.StatusNotFound, msg)
	}

	if len(pathParts) == 4 && strings.HasSuffix(pathParts[3], ".json") {
		p.serveRepodata(wr, req, pathParts[1], pathParts[2], pathParts[3])
		return
	}

	if p.AllowedFilenames != nil && !p.AllowedFilenames.Contains(filePath) {
		msg := "Invalid filepath: " + req.URL.Path
		http.Error(wr, msg, http.StatusNotFound)
		log.Println(logPrefix, http.StatusNotFound, msg)
		return
	}

	client := &http.Client{Timeout: time.Duration(p.Cfg.TimeoutSeconds) * time.Second}

	condaUrl := p.Cfg.CondaHost + req.URL.Path
	log.Println("Fetching:", condaUrl)

	resp, err := client.Get(condaUrl)
	if err != nil {
		http.Error(wr, "Server Error", http.StatusInternalServerError)
		log.Println(logPrefix, http.StatusInternalServerError, "ServeHTTP:", err)
		return
	}
	defer resp.Body.Close()

	log.Println(logPrefix, resp.Status)

	copyHeader(wr.Header(), resp.Header)
	wr.WriteHeader(resp.StatusCode)
	if _, err := io.Copy(wr, resp.Body); err != nil {
		log.Println(logPrefix, "ERROR ServeHTTP:", err)
		return
	}
}

func main() {
	configFile := flag.String("cfg", "", "Configuration file")
	flag.Parse()

	if *configFile == "" {
		log.Fatalf("Configuration file required")
	}

	cfg, err := repodata.LoadCondaRepoConfig(*configFile)
	if err != nil {
		log.Fatalf("Failed to load configuration file: %s", err)
	}

	allowedFilelistName := filepath.Join(cfg.FilteredRepodataDir, "filenames.txt")
	allowedFilenames := repodata.ParseListFromFile(allowedFilelistName)
	log.Printf("Loaded %d allowed filenames from %s", allowedFilenames.Len(), allowedFilelistName)

	// https://blog.cloudflare.com/the-complete-guide-to-golang-net-http-timeouts/
	srv := &http.Server{
		ReadTimeout:  time.Duration(cfg.TimeoutSeconds) * time.Second,
		WriteTimeout: time.Duration(cfg.TimeoutSeconds) * time.Second,
	}

	srv.Handler = &proxy{
		AllowedFilenames: allowedFilenames,
		Cfg:              cfg,
	}
	srv.Addr = cfg.Listen

	log.Println("Starting conda-proxy server on", cfg.Listen)
	if err := srv.ListenAndServe(); err != nil {
		log.Fatal("ListenAndServe:", err)
	}
}
