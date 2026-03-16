package ssh

import (
	"fmt"
	"os"
	"path"

	fs "github.com/sdsc-ordes/quitsh/pkg/filesystem"
)

// AddKnownHost adds a host name `hostName` with key `publicKey` to the known hosts SSH file.
func AddKnownHost(hostName string, publicKey string) error {
	dir, e := os.UserHomeDir()
	if e != nil {
		return e
	}

	f, e := os.OpenFile(path.Join(dir, ".ssh/known_hosts"),
		os.O_WRONLY|os.O_CREATE|os.O_APPEND,
		fs.DefaultPermissionsFile)
	if e != nil {
		return e
	}

	defer func() { _ = f.Close() }()

	_, e = fmt.Fprintf(f, "\n%s %s", hostName, publicKey)

	return e
}
