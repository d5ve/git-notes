package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
)

const defaultBranch = "master"

type Repo struct {
	Path   string
	Branch string
}

type Config struct {
	Repos []Repo
}

type ConfigReader interface {
	Read(path string) (*Config, error)
}

type JsonConfigReader struct{}

func (c *JsonConfigReader) Read(path string) (*Config, error) {
	bytes, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var doc map[string]interface{}
	err = json.Unmarshal(bytes, &doc)
	if err != nil {
		return nil, err
	}

	repos, found := doc["repos"]
	if !found {
		return nil, errors.New("Unable to read repos list from " + path)
	}
	var config Config
	if rec, ok := repos.([]interface{}); ok {
		for _, val := range rec {
			switch val := val.(type) {
			case string:
				repo := Repo{Path: val, Branch: defaultBranch}
				config.Repos = append(config.Repos, repo)
			case map[string]interface{}:
				rpath, found := val["path"]
				if !found {
					return nil, fmt.Errorf("Unable to read repo path from %v", val)
				}
				branch, found := val["branch"]
				if !found {
					return nil, fmt.Errorf("Unable to read repo branch from %v", val)
				}
				repo := Repo{Path: rpath.(string), Branch: branch.(string)}
				config.Repos = append(config.Repos, repo)
			default:
				return nil, fmt.Errorf("Unexpected type %T in %v", val, val)
			}
		}
	} else {
		return nil, errors.New("Config doesn't contain a repos list")
	}

	return &config, nil
}
