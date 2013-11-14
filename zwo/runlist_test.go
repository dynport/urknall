package zwo

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestAddCommandToRunlistWithoutError(t *testing.T) {
	r := &Runlist{}
	r.setConfig(&struct{}{})

	e := r.AddCommands(Execute("foo"))
	assert.Nil(t, e)
	assert.Equal(t, len(r.actions), 1)
}

func TestAddCommandsToRunlistWithError(t *testing.T) {
	r := &Runlist{}
	r.setConfig(&struct{}{})

	e := r.AddCommands()
	assert.Error(t, e, "empty command given")
	assert.Equal(t, len(r.actions), 0)
}
