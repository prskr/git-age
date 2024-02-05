package ports

import (
	"io"

	"filippo.io/age"
)

const RecipientsFileName = ".agerecipients"

type FileOpenSealer interface {
	FileOpener
	FileSealer
}

type FileSealer interface {
	CanSeal() bool
	AddRecipients(r ...age.Recipient)
	SealFile(dst io.Writer) (io.WriteCloser, error)
}

type FileOpener interface {
	CanOpen() bool
	AddIdentities(identities ...age.Identity)
	OpenFile(reader io.Reader) (io.Reader, error)
}
