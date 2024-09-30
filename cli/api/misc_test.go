//go:build server_test

package api

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestGetServerInfo(t *testing.T) {
	info, err := UserA.context.GetServerInfo()
	assert.Nil(t, err)
	assert.NotEmpty(t, info.StorageBackend)
}
