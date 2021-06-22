package main

import (
	"git-notes/internal/types"
	"log"
	"time"
)

type Watcher interface {
	Watch(repo types.Repo, channel chan types.Repo)
}

type GitWatcher struct {
	git                    Git
	running                bool
	checkInterval          time.Duration
	delayBeforeFiringEvent time.Duration
	delayAfterFiringEvent  time.Duration
}

func (f *GitWatcher) Stop() {
	f.running = false
}

func (f *GitWatcher) Check(repo types.Repo, channel chan types.Repo) {
	dirty, err := f.git.IsDirty(repo)

	if err != nil {
		log.Printf("Failed to get state. Error: %v", err)
	}

	if dirty {
		log.Printf("Changes have been detected.")
		time.Sleep(f.delayBeforeFiringEvent)
		channel <- repo
		time.Sleep(f.delayAfterFiringEvent)
	}
}

func (f *GitWatcher) Watch(repo types.Repo, channel chan types.Repo) {
	f.running = true
	go func() {
		for f.running {
			time.Sleep(f.checkInterval)
			f.Check(repo, channel)
		}
	}()

}
