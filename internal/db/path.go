package db

import (
	"os"
	"path/filepath"
)

// ResolveDBPath は XDG Base Directory 仕様に従って DBファイルパスを返す。
// $XDG_DATA_HOME が設定されていれば $XDG_DATA_HOME/tacktack/tacktack.db、
// 未設定なら ~/.local/share/tacktack/tacktack.db を返す。
func ResolveDBPath() (string, error) {
	dataHome := os.Getenv("XDG_DATA_HOME")
	if dataHome == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		dataHome = filepath.Join(home, ".local", "share")
	}
	return filepath.Join(dataHome, "tacktack", "tacktack.db"), nil
}
