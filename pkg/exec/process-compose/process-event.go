package processcompose

// pcEvent is the `ProcessStateEvent` is taken from `process-compose` we do not serialize
// only stuff we really need.
// Version: 1.110.
type pcEvent struct {
	State pcState `json:"state"`
}

type pcState struct {
	Name     string `json:"name"`
	Status   string `json:"status"`
	IsReady  string `json:"is_ready"` //nolint:tagliatelle // API.
	Restarts int    `json:"restarts"`
	ExitCode int    `json:"exit_code"` //nolint:tagliatelle // API.
}
