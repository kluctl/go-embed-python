package internal

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	log "github.com/sirupsen/logrus"
)

func StripBinaries(path string, arch string) error {
	var binFiles []string

	err := filepath.Walk(path, func(p string, info os.FileInfo, err error) error {
		relPath, err := filepath.Rel(path, p)
		if err != nil {
			return err
		}
		if info.Mode().IsRegular() && info.Mode()&0111 != 0 {
			b, err := os.ReadFile(p)
			if err != nil {
				return err
			}
			if len(b) >= 4 && b[0] == 0x7F && b[1] == 0x45 && b[2] == 0x4c && b[3] == 0x46 {
				binFiles = append(binFiles, relPath)
			}
		}
		return nil
	})
	if err != nil {
		return err
	}

	log.Infof("stripping %d files in %s", len(binFiles), path)

	script := `set -e
export DEBIAN_FRONTEND=noninteractive
apt update
apt install -y llvm
cd /host
`
	for _, file := range binFiles {
		script += fmt.Sprintf("llvm-strip %s\n", file)
	}

	cmd := exec.Command("docker", "run", "-i", "-v", fmt.Sprintf("%s:/host", path), "-w", "/host", "debian:trixie", "sh", "-c", script)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err = cmd.Run()
	if err != nil {
		return err
	}
	return nil
}
