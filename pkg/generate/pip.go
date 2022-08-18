package main

import (
	"os"
	"os/exec"
)

func pipInstallRequirements(libDir string, requirementsFile string) {
	cmd := exec.Command("pip3", "install", "-r", requirementsFile, "-t", libDir)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	if err != nil {
		panic(err)
	}
}
