package backend

import "io"

type FileBackend interface {
	Save(string, map[string]interface{}, io.Reader) (string, error)
}
