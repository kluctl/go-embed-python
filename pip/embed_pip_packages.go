package pip

import (
	"fmt"
	"github.com/kluctl/go-embed-python/embed_util"
	"github.com/kluctl/go-embed-python/internal"
	"github.com/kluctl/go-embed-python/python"
	"math/rand"
	"os"
)

func CreateEmbeddedPipPackages(requirementsFile string, targetDir string) error {
	name := fmt.Sprintf("pip-%d", rand.Uint32())

	ep, err := python.NewEmbeddedPython(name)
	if err != nil {
		return err
	}
	defer ep.Cleanup()

	pipLib, err := NewPipLib(name)
	if err != nil {
		return err
	}
	defer pipLib.Cleanup()

	ep.AddPythonPath(pipLib.GetExtractedPath())

	return CreateEmbeddedPipPackages2(ep, requirementsFile, targetDir)
}

func CreateEmbeddedPipPackages2(ep *python.EmbeddedPython, requirementsFile string, targetDir string) error {
	tmpDir, err := os.MkdirTemp("", "pip-")
	if err != nil {
		return err
	}
	defer os.RemoveAll(tmpDir)

	err = pipInstall(ep, requirementsFile, tmpDir)
	if err != nil {
		return err
	}

	if internal.Exists(targetDir) {
		err = os.RemoveAll(targetDir)
		if err != nil {
			return err
		}
	}

	err = os.MkdirAll(targetDir, 0o755)
	if err != nil {
		return err
	}

	err = embed_util.CopyForEmbed(targetDir, tmpDir)
	if err != nil {
		return err
	}

	return nil
}

func pipInstall(ep *python.EmbeddedPython, requirementsFile string, targetDir string) error {
	cmd := ep.PythonCmd("-m", "pip", "install", "-r", requirementsFile, "-t", targetDir)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	if err != nil {
		return err
	}
	err = internal.CleanupPythonDir(targetDir, nil)
	if err != nil {
		return err
	}
	return nil
}