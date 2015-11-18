package mongobackend

import (
	"io"

	"github.com/InteractiveLecture/media-service/backend"
	"gopkg.in/mgo.v2"
)

type MongoFileBackend struct {
	ServerAddress string
	DbName        string
	GridFSName    string
}

type LoadType int

const (
	ID LoadType = iota
	NAME
)

func New(address, dbname, fsname string) *MongoFileBackend {
	return &MongoFileBackend{
		ServerAddress: address,
		DbName:        dbname,
		GridFSName:    fsname,
	}
}

func (m MongoFileBackend) Save(fileName string, contentType string, meta map[string]interface{}, reader backend.ReadSeekCloser) (string, error) {
	session, connectionError := mgo.Dial(m.ServerAddress)
	if connectionError != nil {
		return "", connectionError
	}
	defer session.Close()
	id := uuid.NewV4().String()
	if fileName == "" {
		fileName = id
	}
	file, err := session.DB(m.DbName).GridFS(m.GridFSName).Create(fileName)
	if err != nil {
		return "", err
	}
	file.SetId(id)
	file.SetMeta(meta)
	file.SetContentType(contentType)
	defer file.Close()
	_, err = io.Copy(file, reader)
	if err != nil {
		return "", err
	}
	return id, nil
}

func (m MongoFileBackend) LoadByName(fileName string) (backend.ReadSeekCloser, *backend.FileMeta, error) {
	return m.load(fileName, NAME)
}

func (m MongoFileBackend) LoadById(id string) (backend.ReadSeekCloser, *backend.FileMeta, error) {
	return m.load(id, ID)
}

func (m MongoFileBackend) load(id string, loadType LoadType) (backend.ReadSeekCloser, *backend.FileMeta, error) {
	session, err := mgo.Dial(m.ServerAddress)
	if err != nil {
		return nil, nil, err
	}
	defer session.Close()
	var file *mgo.GridFile
	if loadType == ID {
		file, err = session.DB("media-service").GridFS("fs").OpenId(id)
	} else {
		file, err = session.DB("media-service").GridFS("fs").Open(id)
	}

	var meta = make(map[string]interface{})
	file.GetMeta(meta)
	fileMeta := backend.NewMeta(file.Name(), file.ContentType(), file.Size(), meta)
	if err != nil {
		return nil, nil, err
	}
	if err != nil {
		return nil, nil, err
	}
	return file, fileMeta, nil
}
