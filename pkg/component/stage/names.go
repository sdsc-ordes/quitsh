package stage

const (
	Aux Stage = "aux"

	Lint  Stage = "lint"
	Build Stage = "build"
	Test  Stage = "test"

	Coverage Stage = "coverage"

	Image  Stage = "image"
	Deploy Stage = "deploy"
)
