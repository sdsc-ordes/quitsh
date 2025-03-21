package runner

type IRunner interface {
	ID() RegisterID
	Run(ctx IContext) error
}
