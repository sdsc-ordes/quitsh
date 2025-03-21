package ci

import (
	"os"

	"github.com/sdsc-ordes/quitsh/pkg/exec/git"
)

// IsRunning returns `true` if we are currently running in CI.
func IsRunning() bool {
	return os.Getenv("CI") == "true"
}

// SetupGit sets up common Git variables in CI.
// - Sets up Git LFS.
// - Sets `user.email` and `user.name`.
func SetupGit(gitx git.Context, user string, userEmail string) error {
	return gitx.Chain().
		Check("lfs", "install").
		Check("config", "--global", "user.email", user).
		Check("config", "--global", "user.name", userEmail).
		Error()
}
