package main

import (
	"git-notes/internal/types"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// TODO: Define types.Repo{"some-path", "trunk"} here and use in tests.

func TestGitRepoMonitor_StartMonitoring(t *testing.T) {
	var gitRepoMonitor = GitRepoMonitor{
		scheduledUpdateInterval: time.Minute,
	}
	var watcher = MockWatcher{}
	var git = MockGit{}

	gitRepoMonitor.StartMonitoring(types.Repo{"some-path", "trunk"}, &watcher, &git)

	assert.Equal(t, types.Repo{"some-path", "trunk"}, watcher.repo)
	assert.Equal(t, 1, git.Count)

	watcher.channel <- watcher.repo

	time.Sleep(1 * time.Second)
	assert.Equal(t, 2, git.Count)
}

func TestGitRepoMonitor_StartMonitoringAutomaticScheduleUpdate(t *testing.T) {
	var gitRepoMonitor = GitRepoMonitor{
		scheduledUpdateInterval: 100 * time.Millisecond,
	}
	var watcher = MockWatcher{}
	var git = MockGit{}

	gitRepoMonitor.StartMonitoring(types.Repo{"some-path", "trunk"}, &watcher, &git)

	assert.Eventually(t, func() bool {
		return git.Count >= 2
	}, 1*time.Second, 10*time.Millisecond)
}

func TestGitRepoMonitor_ScheduleUpdate(t *testing.T) {
	var gitRepoMonitor = GitRepoMonitor{
		scheduledUpdateInterval: 100 * time.Millisecond,
	}

	var channel = make(chan types.Repo)
	var repo types.Repo

	go func() {
		repo = <-channel
	}()

	gitRepoMonitor.scheduleUpdate(types.Repo{"some-path", "trunk"}, channel)

	assert.Eventually(t, func() bool {
		return repo == types.Repo{"some-path", "trunk"}
	}, 1*time.Second, 10*time.Millisecond)
}

type MockWatcher struct {
	repo    types.Repo
	channel chan types.Repo
}

func (m *MockWatcher) Watch(repo types.Repo, channel chan types.Repo) {
	m.repo = repo
	m.channel = channel
}

type MockGit struct {
	Count int
}

func (m *MockGit) IsDirty(repo types.Repo) (bool, error) {
	return false, nil
}

func (m *MockGit) Sync(repo types.Repo) error {
	m.Count++
	return nil
}

func (m *MockGit) Update(repo types.Repo) error {
	return nil
}

func (m *MockGit) GetState(repo types.Repo) (State, error) {
	return Sync, nil
}
