package stage

const (
	Aux Stage = "aux"

	Lint     Stage = "lint"
	Build    Stage = "build"
	Test     Stage = "test"
	Coverage Stage = "coverage"

	Manifest Stage = "manifest"
	Image    Stage = "image"
	Deploy   Stage = "deploy"
)

func AllStages() []Stage {
	return []Stage{
		Aux, Lint, Build, Test, Coverage, Manifest, Image, Deploy,
	}
}
