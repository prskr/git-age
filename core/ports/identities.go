package ports

type Identities interface {
	Generate(comment string) (pubKey string, err error)
}
