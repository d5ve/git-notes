package main

import (
	"git-notes/internal/test_helpers"
	"git-notes/internal/types"
	"log"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

type listener struct {
	repos []types.Repo
}

func setup() (*GitWatcher, *listener, types.Repo, chan types.Repo) {
	var channel chan types.Repo = make(chan types.Repo)

	var watcher = GitWatcher{
		git:                    &GitCmd{},
		running:                false,
		checkInterval:          10 * time.Millisecond,
		delayBeforeFiringEvent: 0,
		delayAfterFiringEvent:  1 * time.Second,
	}

	var repo = test_helpers.SetupGitRepo("watcher", false)

	var listener listener

	go func() {
		for {
			repo = <-channel
			listener.repos = append(listener.repos, repo)
		}
	}()

	return &watcher, &listener, repo, channel
}

func cleanup(watcher *GitWatcher, repo types.Repo) {
	err := os.RemoveAll(repo.Path)
	if err != nil {
		log.Fatalf("Unable to remove %s. Error: %v", repo.Path, err)
	}

	watcher.Stop()
}

func commit(t *testing.T, repo types.Repo) {
	test_helpers.PerformCmd(t, repo, "git", "add", "--all")
	test_helpers.PerformCmd(t, repo, "git", "commit", "-m", "Test")
}

func TestGitWatcher_Watch(t *testing.T) {
	var watcher, listener, repo, channel = setup()
	defer cleanup(watcher, repo)

	watcher.Watch(repo, channel)

	assert.Equal(t, 0, len(listener.repos))

	test_helpers.WriteFile(t, repo, "test.md", "Watch")
	time.Sleep(1 * time.Second)
	assert.Greater(t, len(listener.repos), 0)
	assert.Equal(t, repo, listener.repos[0])
}

func TestGitWatcher_CreateAndModify(t *testing.T) {
	var watcher, listener, repo, channel = setup()
	defer cleanup(watcher, repo)

	watcher.Check(repo, channel)
	assert.Equal(t, 0, len(listener.repos))

	test_helpers.WriteFile(t, repo, "test.md", "Hello")
	watcher.Check(repo, channel)
	assert.Equal(t, 1, len(listener.repos))
	assert.Equal(t, repo, listener.repos[0])

	commit(t, repo)

	watcher.Check(repo, channel)
	assert.Equal(t, 1, len(listener.repos))
	assert.Equal(t, repo, listener.repos[0])

	test_helpers.WriteFile(t, repo, "test.md", "Hello2")
	watcher.Check(repo, channel)
	assert.Equal(t, 2, len(listener.repos))
	assert.Equal(t, repo, listener.repos[0])
	assert.Equal(t, repo, listener.repos[1])

	commit(t, repo)

	// No change
	test_helpers.WriteFile(t, repo, "test.md", "Hello2")
	watcher.Check(repo, channel)
	assert.Equal(t, 2, len(listener.repos))
}
