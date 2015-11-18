package mongofs

import (
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/satori/go.uuid"

	"gopkg.in/mgo.v2"
)

type Mongofs struct {
	Address    string
	DbName     string
	GridFSName string
}

type FileBackend interface {
	http.FileSystem
	Save(string, io.ReadCloser) (string, error)
}

func New(address, dbname, fsname string) *Mongofs {
	return &Mongofs{address, dbname, fsname}
}

type MongoFile struct {
	mgo.GridFile
	mgo.Session
}

func (m MongoFile) Close() error {
	defer m.Session.Close()
	return m.GridFile.Close()
}

func (m Mongofs) Open(name string) (http.File, error) {
	if strings.HasPrefix(name, "/") {
		name = strings.TrimLeft(name, "/")
	}
	log.Println(name)
	session, err := mgo.Dial(m.Address)
	if err != nil {
		return nil, err
	}
	file, err := session.DB(m.DbName).GridFS(m.GridFSName).Open(name)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	return &MongoFile{*file, *session}, nil
}

func (m MongoFile) Mode() os.FileMode {
	return os.FileMode(0777)
}

func (m MongoFile) ModTime() time.Time {
	return m.UploadDate()
}

func (m MongoFile) IsDir() bool {
	return false
}

func (m MongoFile) Sys() interface{} {
	return nil
}

func (m MongoFile) Stat() (os.FileInfo, error) {
	return &m, nil
}

func (m MongoFile) Readdir(count int) ([]os.FileInfo, error) {
	return nil, errors.New("dirs are not supported")
}

func (m Mongofs) Save(fileExtension string, meta map[string]interface{}, reader io.Reader) (string, error) {
	session, connectionError := mgo.Dial(m.Address)
	if connectionError != nil {
		return "", connectionError
	}
	defer session.Close()
	id := uuid.NewV4().String()
	fileName := fmt.Sprintf("%s.%s", id, fileExtension)
	file, err := session.DB(m.DbName).GridFS(m.GridFSName).Create(fileName)
	file.SetMeta(meta)
	if err != nil {
		return "", err
	}
	defer file.Close()
	_, err = io.Copy(file, reader)
	if err != nil {
		return "", err
	}
	return fileName, nil
}
