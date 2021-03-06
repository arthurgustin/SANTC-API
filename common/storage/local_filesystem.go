package storage

import (
	"context"
	b64 "encoding/base64"
	"errors"
	"io/ioutil"
	"os"
	"path"

	"github.com/Vinubaba/SANTC-API/api/shared"
)

const (
	jpegMimetype = "image/jpeg"
)

var (
	ErrUnsupportedFileFormat = errors.New("for now, only jpeg is supported. the image must have the following pattern: 'data:image/jpeg;base64,[big 64encoded image string]'")
)

type Storage interface {
	Store(ctx context.Context, b64image string, folder string) (string, error)
	Get(ctx context.Context, filename string) (string, error)
	Delete(ctx context.Context, filename string) error
}

type LocalStorage struct {
	Config          *shared.AppConfig `inject:""`
	StringGenerator interface {
		GenerateUuid() string
	} `inject:""`
}

func (s *LocalStorage) Store(ctx context.Context, encodedImage, mimeType string, folder string) (string, error) {
	if mimeType != jpegMimetype {
		return "", ErrUnsupportedFileFormat
	}

	decoded, err := b64.StdEncoding.DecodeString(encodedImage)
	if err != nil {
		return "", err
	}

	id := s.StringGenerator.GenerateUuid()
	var fileName string
	if folder != "" {
		fileName = path.Clean(s.Config.LocalStoragePath + "/" + folder + "/" + id + ".jpg")
	} else {
		fileName = path.Clean(s.Config.LocalStoragePath + "/" + id + ".jpg")
	}

	file, err := os.Create(fileName)
	if err != nil {
		return "", err
	}
	defer file.Close()

	_, err = file.Write(decoded)
	if err != nil {
		return "", err
	}
	file.Sync()

	return fileName, nil
}

func (s *LocalStorage) Get(ctx context.Context, filePath string) (string, error) {
	file, err := ioutil.ReadFile(filePath)
	if err != nil {
		return "", err
	}

	return b64.StdEncoding.EncodeToString(file), nil
}
