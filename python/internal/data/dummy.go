package data

// PLEASE READ THIS!!!!
// This file is really just a dummy. The release process will remove this file and generate some read embedded files
// and commit these into a temporary branch and then tag it. This is to avoid clogging up the main branch with too many
// binary files, which would be a very bad experience when pulling in go-embed-python as a dependency.

func init() {
	panic("You can not use the main branch of go-embed-python as a Go dependency, as this branch does not contain the necessary Python distributions. Please use a tagged release of go-embed-python instead.")
}
