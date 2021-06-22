package test_helpers

import (
	"fmt"
	"git-notes/internal/types"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

type Repos struct {
	Remote types.Repo
	Local  types.Repo
}

func SetupRepos() Repos {
	remote := SetupGitRepo("Remote", true)
	local := SetupGitRepo("Local", false)

	SetupRemote(local, remote)

	log.Printf("Local: %s, Remote: %s", local, remote)
	return Repos{
		Remote: remote,
		Local:  local,
	}
}

func CleanupRepos(repos Repos) {
	err := os.RemoveAll(repos.Remote.Path)
	if err != nil {
		log.Fatalf("Unable to remove %s. Error: %v", repos.Remote.Path, err)
	}

	err = os.RemoveAll(repos.Local.Path)
	if err != nil {
		log.Fatalf("Unable to remove %s. Error: %v", repos.Local.Path, err)
	}
}

func SetupGitRepo(tag string, bare bool) types.Repo {
	path, err := ioutil.TempDir("", fmt.Sprintf("git_test_%s", tag))
	if err != nil {
		log.Fatalf("Unable to create a temp dir for the Remote repo")
	}

	args := []string{"init"}
	if bare {
		args = append(args, "--bare")
	}

	c := exec.Command("git", args...)
	c.Dir = path
	err = c.Run()
	if err != nil {
		log.Fatalf("Unable to init the repo. Path: %v, Error: %v", path, err)
	}

	return types.Repo{
		Path:   path,
		Branch: "trunk",
	}
}

func SetupRemote(local types.Repo, remote types.Repo) {
	c := exec.Command("git", "remote", "add", "origin", remote.Path)
	c.Stderr = os.Stderr
	c.Stdout = os.Stdout
	c.Dir = local.Path
	err := c.Run()
	if err != nil {
		log.Fatalf("Unable to add origin. Error: %v", err)
	}
}

func WriteConfigFile(t *testing.T, path string, filePath string, content string) {
	fullPath := fmt.Sprintf("%s/%s", path, filePath)
	log.Printf("Write file: %v, content: %v", fullPath, content)
	assert.NoError(t, ioutil.WriteFile(fullPath, []byte(content), 0644))
}

func WriteFile(t *testing.T, repo types.Repo, filePath string, content string) {
	fullPath := fmt.Sprintf("%s/%s", repo.Path, filePath)
	log.Printf("Write file: %v, content: %v", fullPath, content)
	assert.NoError(t, ioutil.WriteFile(fullPath, []byte(content), 0644))
}

func PerformCmd(t *testing.T, repo types.Repo, cmd string, args ...string) {
	log.Printf("Run cmd: %v", strings.Join(append([]string{cmd}, args...), " "))
	c := exec.Command(cmd, args...)
	c.Dir = repo.Path
	c.Stdin = os.Stdin
	c.Stdout = os.Stdout
	c.Stderr = os.Stderr
	err := c.Run()
	assert.NoError(t, err)
}
