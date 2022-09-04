package main

import (
	"fmt"
	"github.com/mattn/go-shellwords"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

const PreloadTool string = "importrdf"

// assembleToLoadFolderPath assembles the path to the toLoad folder
// as it is required for the preload tool.
func assembleToLoadFolderPath(repositoryDirectory string) string {
	toLoadFolder := filepath.Join(repositoryDirectory, "toLoad")
	if dExists(toLoadFolder) {
		var absToLoadFolder, err = filepath.Abs(toLoadFolder)
		if err == nil {
			return absToLoadFolder
		} else {
			WarningLogger.Printf("couldn't create absolute path to '%s': %s",
				toLoadFolder, err.Error())
		}
	} else {
		WarningLogger.Printf("couldn't find data to load in '%s' for '%s'",
			toLoadFolder, repositoryDirectory)
	}
	p := "/tmp/graphdb/toLoad"
	err := os.MkdirAll(p, os.ModeDir)
	if err != nil {
		WarningLogger.Printf("couldn't create temporary folder '%s': %s",
			p, err.Error())
	}
	return p
}

// constructArgs constructs the arguments for the preload tool. It returns the
// corresponding arguments, or an error, if the construction failed.
func constructArgs(repositoryDirectory string, toLoadFolder string) ([]string, error) {
	configPath := filepath.Join(repositoryDirectory, "config.ttl")
	if !fExists(configPath) {
		return nil, fmt.Errorf("config.ttl is missing for '%s', but it is required for initializing a repository",
			repositoryDirectory)
	}
	absConfigPath, err := filepath.Abs(configPath)
	if err != nil {
		return nil, fmt.Errorf("abosulute path to config.ttl couldn't be built for '%s': %s",
			repositoryDirectory, err.Error())
	}
	args := []string{
		"load",
		"-c",
		absConfigPath,
	}
	preloadCommandArgs := os.Getenv("PRELOAD_ARGS")
	if preloadCommandArgs != "" {
		envArgs, err := shellwords.Parse(preloadCommandArgs)
		if err != nil {
			return nil, fmt.Errorf("the command line arguments of 'PRELOAD_ARGS' couldn't be parsed: %s",
				err.Error())
		}
		args = append(args, envArgs...)
	} else {
		args = append(args, "--partial-load", "--force")
	}
	args = append(args, toLoadFolder)
	return args, nil
}

// cleanTemporaryLoadFolder cleans the temporary folder, if it has been created.
func cleanTemporaryLoadFolder(toLoadFolder string) {
	if toLoadFolder == "/tmp/graphdb/toLoad" {
		err := os.RemoveAll("/tmp/graphdb/toLoad")
		if err != nil {
			WarningLogger.Printf("couldn't delete temporary 'toLoad' folder: %s",
				err.Error())
		}
	}
}

// InitRepository initializes the repository configured in the given directory.
// Returns true, if the repository could be initialized, otherwise false.
func InitRepository(repositoryDirectory string) error {
	InfoLogger.Printf("----- CHECK %s. -----\n", repositoryDirectory)
	if fExists(filepath.Join(repositoryDirectory, "init.lock")) {
		InfoLogger.Printf("----- %s. ALREADY INITIALIZED -----\n",
			repositoryDirectory)
		return nil
	}
	toLoadFolder := assembleToLoadFolderPath(repositoryDirectory)
	args, err := constructArgs(repositoryDirectory, toLoadFolder)
	if err != nil {
		return err
	}
	cmd := exec.Command(PreloadTool, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err = cmd.Run()
	if err != nil {
		cleanTemporaryLoadFolder(toLoadFolder)
		return fmt.Errorf("execution of '%s %s' failed: %s", PreloadTool,
			strings.Join(args, " "), err.Error())
	}
	err = ioutil.WriteFile(filepath.Join(repositoryDirectory, "init.lock"),
		[]byte("locked"), 0644)
	if err != nil {
		WarningLogger.Printf("failed to write lock file for successful initialization of '%s': %s",
			repositoryDirectory, err.Error())
	}
	cleanTemporaryLoadFolder(toLoadFolder)
	return nil
}
