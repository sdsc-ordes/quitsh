package exec

type CmdContextChain struct {
	ctx *CmdContext
	err error
}

// CmdContextChain returns a command ctx where you can chain commands together.
// The context will fail the chain if any error happens.
func (c *CmdContext) Chain() CmdContextChain {
	return CmdContextChain{ctx: c}
}

// Check executes `Check` over the context and return the chain.
func (c CmdContextChain) Check(args ...string) CmdContextChain {
	if c.err == nil {
		c.err = c.ctx.Check(args...)
	}

	return c
}

// CheckWithEC executes `CheckWithEC` over the context and return the chain.
func (c CmdContextChain) CheckWithEC(handleExit ExitCodeHandler, args ...string) CmdContextChain {
	if c.err == nil {
		c.err = c.ctx.CheckWithEC(handleExit, args...)
	}

	return c
}

// Get executes `Get` over the context and returns the result.
func (c CmdContextChain) Get(args ...string) (string, error) {
	if c.err == nil {
		return c.ctx.Get(args...)
	}

	return "", c.err
}

// GetWithEC executes `GetWithEC` over the context and returns the result.
func (c CmdContextChain) GetWithEC(handleExit ExitCodeHandler, args ...string) (string, error) {
	if c.err == nil {
		return c.ctx.GetWithEC(handleExit, args...)
	}

	return "", c.err
}

// GetSplit executes `GetSplit` over the context and returns the result.
func (c CmdContextChain) GetSplit(args ...string) ([]string, error) {
	if c.err == nil {
		return c.ctx.GetSplit(args...)
	}

	return nil, c.err
}

// GetSplitWithEC executes `GetSplitWithEC` over the context and returns the result.
func (c CmdContextChain) GetSplitWithEC(
	handleExit ExitCodeHandler,
	args ...string,
) ([]string, error) {
	if c.err == nil {
		return c.ctx.GetSplitWithEC(handleExit, args...)
	}

	return nil, c.err
}

// Error  returns the error on the chain.
func (c CmdContextChain) Error() error {
	return c.err
}
