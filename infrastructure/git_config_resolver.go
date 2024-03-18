package infrastructure

import "github.com/alecthomas/kong"

var _ kong.Resolver = (*GitConfigResolver)(nil)

type GitConfigResolver struct{}

func (g GitConfigResolver) Validate(app *kong.Application) error {
	return nil
}

func (g GitConfigResolver) Resolve(context *kong.Context, parent *kong.Path, flag *kong.Flag) (interface{}, error) {
	// TODO implement me
	panic("implement me")
}
