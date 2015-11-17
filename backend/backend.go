package backend

import "io"

type FileBackend interface {
	Save(string, string, map[string]interface{}, ReadSeekCloser) (string, error)
	LoadByName(string) (ReadSeekCloser, *FileMeta, error)
	LoadById(string) (ReadSeekCloser, *FileMeta, error)
}

type FileMeta struct {
	Size        int64
	Name        string
	ContentType string
	Meta        map[string]interface{}
}

type ReadSeekCloser interface {
	io.Reader
	io.Seeker
	io.Closer
}

func NewMeta(name, contentType string, size int64, meta map[string]interface{}) *FileMeta {
	return &FileMeta{
		size,
		name,
		contentType,
		meta,
	}
}
