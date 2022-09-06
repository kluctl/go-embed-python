package pip

import (
	"fmt"
	"github.com/kluctl/go-embed-python/embed_util"
	"github.com/kluctl/go-embed-python/internal"
	"github.com/kluctl/go-embed-python/python"
	"math/rand"
	"os"
	"path/filepath"
)

func CreateEmbeddedPipPackagesForKnownPlatforms(requirementsFile string, targetDir string) error {
	platforms := map[string][]string{
		"darwin-amd64":  {"macosx_11_0_x86_64"},
		"darwin-arm64":  {"macosx_11_0_arm64"},
		"linux-amd64":   {"manylinux_2_28_x86_64", "manylinux2014_x86_64"},
		"linux-arm64":   {"manylinux_2_28_aarch64", "manylinux2014_aarch64"},
		"windows-amd64": {"win_amd64"},
	}

	for goPlatform, pipPlatforms := range platforms {
		for i, pipPlatform := range pipPlatforms {
			err := CreateEmbeddedPipPackages("requirements.txt", pipPlatform, filepath.Join(targetDir, goPlatform))
			if err != nil {
				if i == len(pipPlatforms)-1 {
					return err
				}
			}
		}
	}
	return nil
}

func CreateEmbeddedPipPackages(requirementsFile string, platform string, targetDir string) error {
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

	return CreateEmbeddedPipPackages2(ep, requirementsFile, platform, targetDir)
}

func CreateEmbeddedPipPackages2(ep *python.EmbeddedPython, requirementsFile string, platform string, targetDir string) error {
	tmpDir, err := os.MkdirTemp("", "pip-")
	if err != nil {
		return err
	}
	defer os.RemoveAll(tmpDir)

	err = pipInstall(ep, requirementsFile, platform, tmpDir)
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

func pipInstall(ep *python.EmbeddedPython, requirementsFile string, platform string, targetDir string) error {
	args := []string{"-m", "pip", "install", "-r", requirementsFile, "-t", targetDir}
	if platform != "" {
		args = append(args, "--platform", platform, "--only-binary=:all:")
	}

	cmd := ep.PythonCmd(args...)
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
