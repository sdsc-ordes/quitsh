package fs

import "os"

const (
	DefaultPermissionsDir  = os.FileMode(0775)
	DefaultPermissionsFile = os.FileMode(0664)
	StrictPermissionsDir   = os.FileMode(0700)
	StrictPermissionsFile  = os.FileMode(0600)
)
