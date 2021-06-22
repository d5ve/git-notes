package main

import (
	"log"
	"time"
)

type PathMonitor interface {
	StartMonitoring(repo Repo, watcher Watcher, git Git)
	scheduleUpdate(repo Repo, channel chan Repo)
}

type GitRepoMonitor struct {
	scheduledUpdateInterval time.Duration
}

func (g *GitRepoMonitor) scheduleUpdate(repo Repo, channel chan Repo) {
	time.AfterFunc(g.scheduledUpdateInterval, func() {
		channel <- repo
		g.scheduleUpdate(repo, channel)
	})
}

func (g *GitRepoMonitor) StartMonitoring(repo Repo, watcher Watcher, git Git) {
	var channel = make(chan Repo)
	err := git.Sync(repo)
	if err != nil {
		log.Printf("Syncing failed. Err: %v", err)
	}
	g.scheduleUpdate(repo, channel)

	watcher.Watch(repo, channel)

	go func() {
		for {
			path := <-channel
			err = git.Sync(path)
			if err != nil {
				log.Printf("Syncing failed. Err: %v", err)
			}
		}
	}()

	log.Printf("Git notes is monitoring %s:%s", repo.Path, repo.Branch)
}
