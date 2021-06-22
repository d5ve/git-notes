package main

import (
	"git-notes/internal/types"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestJsonConfigReader_Read(t *testing.T) {
	reader := JsonConfigReader{}
	config, err := reader.Read("./git-notes.json.example")
	assert.NoError(t, err)

	assert.Equal(t, []types.Repo{types.Repo{"/Users/ash/todos", "trunk"}, types.Repo{"/Users/ash/projects/personal-notes", "master"}}, config.Repos)
}
