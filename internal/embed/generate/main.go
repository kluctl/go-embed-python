package main

import (
	"fmt"
	"github.com/gobwas/glob"
	"github.com/klauspost/compress/zstd"
	"github.com/kluctl/go-embed-python/embed_util"
	"github.com/kluctl/go-embed-python/internal"
	log "github.com/sirupsen/logrus"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"sync"
)

// versions taken from https://github.com/indygreg/python-build-standalone/releases/
const (
	pythonVersionBase       = "3.10"
	pythonVersionFull       = "3.10.6"
	pythonStandaloneVersion = "20220802"
)

var pythonDists = map[string]string{
	"linux":   "unknown-linux-gnu-lto-full",
	"darwin":  "apple-darwin-lto-full",
	"windows": "pc-windows-msvc-shared-pgo-full",
}

var archMapping = map[string]string{
	"amd64": "x86_64",
	"386":   "i686",
	"arm64": "aarch64",
}

var removeLibs = []string{
	"distutils",
	"ensurepip",
	"idlelib",
	"lib2to3",
	"multiprocessing",
	"pydoc_data",
	"site-packages",
	"sqlite3",
	"test",
	"turtledemo",
	"venv",
	"bin", // not really a library, but erroneously installed by jsonpath_ng
}

var keepNixPatterns = []glob.Glob{
	glob.MustCompile("bin/**"),
	glob.MustCompile("lib/*.so*"),
	glob.MustCompile("lib/*.dylib"),
	glob.MustCompile("lib/python3.*/**"),
}
var keepWinPatterns = []glob.Glob{
	glob.MustCompile("Lib/**"),
	glob.MustCompile("DLLs/**"),
	glob.MustCompile("*.dll"),
	glob.MustCompile("*.exe"),
}

var downloadLock sync.Mutex

func main() {
	var wg sync.WaitGroup

	type job struct {
		os           string
		arch         string
		out          string
		keepPatterns []glob.Glob
	}
	jobs := []job{
		{"linux", "amd64", "data/python-linux-amd64", keepNixPatterns},
		{"linux", "arm64", "data/python-linux-arm64", keepNixPatterns},
		{"darwin", "amd64", "data/python-darwin-amd64", keepNixPatterns},
		{"darwin", "arm64", "data/python-darwin-arm64", keepNixPatterns},
		{"windows", "amd64", "data/python-windows-amd64", keepWinPatterns},
	}
	for _, j := range jobs {
		j := j
		wg.Add(1)
		go func() {
			downloadAndCopy(j.os, j.arch, j.out, j.keepPatterns)
			wg.Done()
		}()
	}
	wg.Wait()
}

func downloadAndCopy(osName string, arch string, out string, keepPatterns []glob.Glob) {
	dist, ok := pythonDists[osName]
	if !ok {
		log.Panicf("no dist for %s", osName)
	}

	downloadPath := download(osName, arch, dist)

	extractPath := downloadPath + ".extracted"
	err := os.RemoveAll(extractPath)
	if err != nil {
		log.Panic(err)
	}

	extract(downloadPath, extractPath)

	installPath := filepath.Join(extractPath, "python", "install")

	var libPath string
	if osName == "windows" {
		libPath = filepath.Join(installPath, "Lib")
	} else {
		libPath = filepath.Join(installPath, "lib", fmt.Sprintf("python%s", pythonVersionBase))
	}

	for _, lib := range removeLibs {
		_ = os.RemoveAll(filepath.Join(libPath, lib))
	}

	err = internal.CleanupPythonDir(installPath, keepPatterns)
	if err != nil {
		panic(err)
	}

	copyToEmbedDir(out, installPath)
}

func copyToEmbedDir(out string, dir string) {
	if internal.Exists(out) {
		err := os.RemoveAll(out)
		if err != nil {
			log.Panic(err)
		}
	}
	err := os.Mkdir(out, 0o755)
	if err != nil {
		log.Panic(err)
	}

	err = embed_util.CopyForEmbed(out, dir)
	if err != nil {
		log.Panic(err)
	}
}

func download(osName, arch, dist string) string {
	downloadLock.Lock()
	defer downloadLock.Unlock()

	pythonArch, ok := archMapping[arch]
	if !ok {
		log.Errorf("arch %s not supported", arch)
		os.Exit(1)
	}
	fname := fmt.Sprintf("cpython-%s+%s-%s-%s.tar.zst", pythonVersionFull, pythonStandaloneVersion, pythonArch, dist)
	downloadPath := filepath.Join(internal.GetTmpBaseDir(), "python-download", fname)
	downloadUrl := fmt.Sprintf("https://github.com/indygreg/python-build-standalone/releases/download/%s/%s", pythonStandaloneVersion, fname)

	if _, err := os.Stat(downloadPath); err == nil {
		log.Infof("skipping download of %s", downloadUrl)
		return downloadPath
	}

	err := os.MkdirAll(filepath.Dir(downloadPath), 0o755)
	if err != nil {
		log.Errorf("mkdirs failed: %v", err)
		os.Exit(1)
	}
	log.Infof("downloading %s", downloadUrl)

	r, err := http.Get(downloadUrl)
	if err != nil {
		log.Errorf("download failed: %v", err)
		os.Exit(1)
	}
	if r.StatusCode == http.StatusNotFound {
		log.Errorf("404 not found")
		os.Exit(1)
	}
	defer r.Body.Close()

	fileData, err := io.ReadAll(r.Body)

	err = os.WriteFile(downloadPath, fileData, 0o640)
	if err != nil {
		log.Errorf("writing file failed: %v", err)
		os.Remove(downloadPath)
		os.Exit(1)
	}

	return downloadPath
}

func extract(archivePath string, targetPath string) string {
	f, err := os.Open(archivePath)
	if err != nil {
		log.Errorf("opening file failed: %v", err)
		os.Exit(1)
	}
	defer f.Close()

	z, err := zstd.NewReader(f)
	if err != nil {
		log.Errorf("decompression failed: %v", err)
		os.Exit(1)
	}
	defer z.Close()

	log.Infof("decompressing %s", archivePath)
	err = internal.ExtractTarStream(z, targetPath)
	if err != nil {
		log.Errorf("decompression failed: %v", err)
		os.Exit(1)
	}

	return targetPath
}
