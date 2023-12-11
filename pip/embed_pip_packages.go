package pip

import (
	"fmt"
	"github.com/kluctl/go-embed-python/embed_util"
	"github.com/kluctl/go-embed-python/internal"
	"github.com/kluctl/go-embed-python/python"
	"math/rand"
	"os"
	"path/filepath"
	"strings"
)

func CreateEmbeddedPipPackagesForKnownPlatforms(requirementsFile string, targetDir string) error {
	platforms := map[string][]string{
		"darwin-amd64":  {"macosx_11_0_x86_64", "macosx_12_0_x86_64"},
		"darwin-arm64":  {"macosx_11_0_arm64", "macosx_12_0_arm64"},
		"linux-amd64":   {"manylinux_2_17_x86_64", "manylinux_2_28_x86_64", "manylinux2014_x86_64"},
		"linux-arm64":   {"manylinux_2_17_aarch64", "manylinux_2_28_aarch64", "manylinux2014_aarch64"},
		"windows-amd64": {"win_amd64"},
	}

	for goPlatform, pipPlatforms := range platforms {
		s := strings.Split(goPlatform, "-")
		goOs, goArch := s[0], s[1]
		err := CreateEmbeddedPipPackages(requirementsFile, goOs, goArch, pipPlatforms, targetDir)
		if err != nil {
			return err
		}
	}
	return nil
}

func CreateEmbeddedPipPackages(requirementsFile string, goOs string, goArch string, pipPlatforms []string, targetDir string) error {
	name := fmt.Sprintf("pip-%d", rand.Uint32())

	// ensure we have a stable extract path for the python distribution (otherwise shebangs won't be stable)
	tmpDir := filepath.Join("/tmp", fmt.Sprintf("python-pip-%s-%s", goOs, goArch))
	ep, err := python.NewEmbeddedPythonWithTmpDir(tmpDir, false)
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

	return CreateEmbeddedPipPackages2(ep, requirementsFile, goOs, goArch, pipPlatforms, targetDir)
}

func CreateEmbeddedPipPackages2(ep *python.EmbeddedPython, requirementsFile string, goOs string, goArch string, pipPlatforms []string, targetDir string) error {
	tmpDir, err := os.MkdirTemp("", "pip-")
	if err != nil {
		return err
	}
	defer os.RemoveAll(tmpDir)

	err = pipInstall(ep, requirementsFile, pipPlatforms, tmpDir)
	if err != nil {
		return err
	}

	err = os.MkdirAll(targetDir, 0o755)
	if err != nil {
		return err
	}

	platformTargetDir := targetDir
	if goOs != "" {
		platformTargetDir = filepath.Join(platformTargetDir, fmt.Sprintf("%s-%s", goOs, goArch))
	}

	if internal.Exists(platformTargetDir) {
		err = os.RemoveAll(platformTargetDir)
		if err != nil {
			return err
		}
	}

	err = os.Mkdir(platformTargetDir, 0o755)
	if err != nil {
		return err
	}

	err = embed_util.CopyForEmbed(platformTargetDir, tmpDir)
	if err != nil {
		return err
	}

	err = embed_util.WriteEmbedGoFile(targetDir, goOs, goArch)
	if err != nil {
		return err
	}

	return nil
}

func pipInstall(ep *python.EmbeddedPython, requirementsFile string, platforms []string, targetDir string) error {
	args := []string{"-m", "pip", "install", "-r", requirementsFile, "-t", targetDir}
	if len(platforms) != 0 {
		for _, p := range platforms {
			args = append(args, "--platform", p)
		}
		args = append(args, "--only-binary=:all:")
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
