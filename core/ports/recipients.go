package ports

type Recipients interface {
	Append(pubKey string, comment string) error
}
