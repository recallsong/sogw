package core

func (a *Api) EvalLambda(ctx *RequestContext, text string) error {
	// not implement yet
	ctx.ReqCtx.WriteString("lambda expression is not implement yet.")
	return nil
}
