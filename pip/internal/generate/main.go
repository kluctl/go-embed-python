package main

import (
	"fmt"
	go_embed_python "github.com/kluctl/go-embed-python"
	"github.com/kluctl/go-embed-python/embed_util"
	"github.com/kluctl/go-embed-python/internal"
	"io"
	"math/rand"
	"net/http"
	"os"
)

func main() {
	rndName := fmt.Sprintf("pip-install-%d", rand.Uint32())
	ep, err := go_embed_python.NewEmbeddedPython(rndName)
	if err != nil {
		panic(err)
	}
	defer ep.Cleanup()

	bootstrapPip(ep)

	tmpDir, err := os.MkdirTemp("", "pip-")
	if err != nil {
		panic(err)
	}
	defer os.RemoveAll(tmpDir)

	installPip(ep, tmpDir)

	targetDir := "./data/pip"
	if internal.Exists(targetDir) {
		err = os.RemoveAll(targetDir)
		if err != nil {
			panic(err)
		}
	}

	err = os.MkdirAll(targetDir, 0o755)
	if err != nil {
		panic(err)
	}

	err = embed_util.CopyForEmbed(targetDir, tmpDir)
	if err != nil {
		panic(err)
	}
}

func bootstrapPip(ep *go_embed_python.EmbeddedPython) {
	getPip := downloadGetPip()
	defer os.Remove(getPip)

	cmd := ep.PythonCmd(getPip)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	if err != nil {
		panic(err)
	}
}

func installPip(ep *go_embed_python.EmbeddedPython, targetDir string) {
	cmd := ep.PythonCmd("-m", "pip", "install", "-r", "requirements.txt", "-t", targetDir)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	if err != nil {
		panic(err)
	}
	err = internal.CleanupPythonDir(targetDir, nil)
	if err != nil {
		panic(err)
	}
}

func downloadGetPip() string {
	resp, err := http.Get("https://bootstrap.pypa.io/get-pip.py")
	if err != nil {
		panic(err)
	}
	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		panic("failed to download get-pip.py: " + resp.Status)
	}

	tmpFile, err := os.CreateTemp("", "get-pip.py")
	if err != nil {
		panic(err)
	}
	defer tmpFile.Close()

	_, err = io.Copy(tmpFile, resp.Body)
	if err != nil {
		os.Remove(tmpFile.Name())
		panic(err)
	}

	return tmpFile.Name()
}
