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
	dir = path.Join(dir, ".ssh")
	file := path.Join(dir, "known_hosts")

	e = os.MkdirAll(dir, fs.StrictPermissionsDir)
	if e != nil {
		return e
	}

	f, e := os.OpenFile(file,
		os.O_WRONLY|os.O_CREATE|os.O_APPEND,
		fs.StrictPermissionsFile)
	if e != nil {
		return e
	}

	defer func() { _ = f.Close() }()

	_, e = fmt.Fprintf(f, "\n%s %s", hostName, publicKey)

	return e
}
