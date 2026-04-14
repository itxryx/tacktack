package db

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestResolveDBPath_XDGDataHome(t *testing.T) {
	t.Setenv("XDG_DATA_HOME", "/tmp/testxdg")
	path, err := ResolveDBPath()
	require.NoError(t, err)
	assert.Equal(t, "/tmp/testxdg/tacktack/tacktack.db", path)
}

func TestResolveDBPath_Default(t *testing.T) {
	t.Setenv("XDG_DATA_HOME", "")
	path, err := ResolveDBPath()
	require.NoError(t, err)

	home, err := os.UserHomeDir()
	require.NoError(t, err)

	expected := filepath.Join(home, ".local", "share", "tacktack", "tacktack.db")
	assert.Equal(t, expected, path)
	assert.True(t, strings.HasSuffix(path, "/tacktack/tacktack.db"))
}
