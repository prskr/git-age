package dto

type GenerateIdentityCommand struct {
	Comment string
	Remote  string
}

type IdentitiesQuery struct {
	Remotes []string
}
