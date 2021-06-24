package main

import (
	"fmt"
	"git-notes/internal/types"
	"log"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"time"
)

const (
	Error     State = "error"
	Dirty     State = "dirty"
	Ahead     State = "ahead"
	OutOfSync State = "out-of-sync"
	Sync      State = "sync"
)

type State string

type Git interface {
	IsDirty(repo types.Repo) (bool, error)
	GetState(repo types.Repo) (State, error)
	Sync(repo types.Repo) error
	Update(repo types.Repo) error
}

type GitCmd struct {
}

func (g *GitCmd) Sync(repo types.Repo) error {
	state, err := g.GetState(repo)
	log.Printf("Starting state: %s", state)
	if err != nil {
		return fmt.Errorf("performing GetState() failed. Err: %v", err)
	}

	for {
		if state == Sync {
			return nil
		}

		err = g.Update(repo)
		if err != nil {
			return fmt.Errorf("performing Update() failed. Err: %v", err)
		}
		nextState, err := g.GetState(repo)
		if err != nil {
			return fmt.Errorf("performing GetState() failed. Err: %v", err)
		}
		log.Printf("Next state: %s", nextState)

		if state == nextState {
			return fmt.Errorf("state doesn't change. Something is wrong")
		}

		state = nextState
	}
}

func runCmd(repo types.Repo, command string, args ...string) (string, error) {
	cmd := exec.Command(command, args...)
	cmd.Dir = repo.Path

	out, err := cmd.CombinedOutput()
	return string(out), err
}

func (g *GitCmd) IsDirty(repo types.Repo) (bool, error) {
	out, err := runCmd(repo, "git", "status", "--porcelain")
	if err != nil {
		return false, fmt.Errorf("unable to get status. Error: %v", err)
	}
	out = strings.TrimSpace(string(out))
	//fmt.Println("IsDirty(" + repo.Path + ") = [" + out + "]")
	dirty := out != ""
	return dirty, nil
}

func (g *GitCmd) GetState(repo types.Repo) (State, error) {
	log.Printf("Computing the state of %s", repo.Path)

	dirty, err := g.IsDirty(repo)
	if err != nil {
		return Error, fmt.Errorf("unable to get status. Error: %v", err)
	}
	if dirty {
		return Dirty, nil
	} else {
		state, err := GetStateAgainstRemote(repo)
		if err != nil {
			return Error, err
		}
		return state, nil
	}
}

func ParseStatusBranch(repo types.Repo, status string) (State, error) {
	// 5 variants of status branch
	// ## $branch
	// ## $branch...origin/$branch
	// ## $branch...origin/$branch [ahead 1]
	// ## $branch...origin/$branch [behind 1]
	// ## $branch...origin/$branch [ahead 1, behind 1]

	pat := fmt.Sprintf("## %s(\\.\\.\\.origin\\/%s *(\\[(ahead|behind) *[0-9]+ *(, *behind *[0-9]+)? *])?)?", repo.Branch, repo.Branch)
	reg := regexp.MustCompile(pat)
	matches := reg.FindAllStringSubmatch(status, -1)

	if len(matches) == 0 {
		return Error, fmt.Errorf("unable to parse status: %v", status)
	}

	groups := matches[0]
	if groups[0] == "" {
		return Error, fmt.Errorf("unable to parse status: %v", status)
	}

	// ## $branch
	if groups[1] == "" {
		return Ahead, nil
	}

	// ## $branch...origin/$branch
	if groups[2] == "" {
		return Sync, nil
	}

	// ## $branch...origin/$branch
	if groups[3] == "behind" {
		return OutOfSync, nil
	}

	// ## $branch...origin/$branch
	if groups[3] == "ahead" {
		if groups[4] == "" {
			return Ahead, nil
		} else {
			return OutOfSync, nil
		}
	}

	return Error, fmt.Errorf("unable to parse status: %v of repo: %s:%s", status, repo.Path, repo.Branch)
}

func GetStateAgainstRemote(repo types.Repo) (State, error) {
	_, err := runCmd(repo, "git", "fetch")
	if err != nil {
		return Error, fmt.Errorf("unable to fetch. Error: %v", err)
	}

	status, err := runCmd(repo, "git", "status", "--branch", "--porcelain")
	if err != nil {
		return Error, fmt.Errorf("unable to fetch. Error: %v", err)
	}

	status = strings.TrimSpace(string(status))
	//fmt.Println("GetStateAgainstRemote(" + repo.Path + ") = [" + status + "]")

	return ParseStatusBranch(repo, status)
}

func (g *GitCmd) Update(repo types.Repo) error {
	state, err := g.GetState(repo)

	if err != nil {
		return err
	}

	switch state {
	case Error:
	case Dirty:
		err = AddAndCommit(repo)
	case Ahead:
		err = Push(repo)
	case OutOfSync:
		err = Merge(repo)
	case Sync:
	}

	return err
}

func AddAndCommit(repo types.Repo) error {
	err := Add(repo)
	if err != nil {
		return err
	}
	return Commit(repo)
}

func Merge(repo types.Repo) error {
	cmd := exec.Command("git", "merge", "origin/"+repo.Branch, "--allow-unrelated-histories", "--no-commit")
	cmd.Dir = repo.Path
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	_ = cmd.Run() // Merge fails if there's conflict. So, we ignore the failure.
	return nil
}

func Push(repo types.Repo) error {
	cmd := exec.Command("git", "push", "origin", repo.Branch, "-u")
	cmd.Dir = repo.Path
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func Add(repo types.Repo) error {
	cmd := exec.Command("git", "add", "--all")
	cmd.Dir = repo.Path
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func Commit(repo types.Repo) error {
	cmd := exec.Command("git", "-c", "user.name='Git notes'", "-c", "user.email='git-notes@noemail.com'", "commit", "-m", fmt.Sprintf("Commited at %v", time.Now()))
	cmd.Dir = repo.Path
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func NewGoGit() GitCmd {
	return GitCmd{}
}
