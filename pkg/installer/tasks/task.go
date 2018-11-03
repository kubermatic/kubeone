package tasks

type Task interface {
	Execute(ctx *Context) error
}
