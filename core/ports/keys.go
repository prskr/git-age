package ports

import (
	"errors"
	"io"
	"os"
)

const RecipientsFileName = ".agerecipients"

type FileOpenSealer interface {
	FileOpener
	FileSealer
}

type FileSealer interface {
	SealFile(dst io.Writer) (io.WriteCloser, error)
}

type FileOpener interface {
	OpenFile(reader io.Reader) (io.Reader, error)
}

func SealFileAtTo(sealer FileSealer, srcPath, dstPath string) (err error) {
	src, err := os.Open(srcPath)
	if err != nil {
		return err
	}

	defer func() {
		_ = src.Close()
	}()

	dst, err := os.Create(dstPath)
	if err != nil {
		return err
	}

	defer func() {
		err = errors.Join(err, dst.Close())
	}()

	encryptWriter, err := sealer.SealFile(dst)
	if err != nil {
		return err
	}

	defer func() {
		err = errors.Join(err, encryptWriter.Close())
	}()

	_, err = io.Copy(encryptWriter, src)
	return err
}
