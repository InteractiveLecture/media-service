package mongofs

import (
	"bytes"
	"fmt"
	"mime"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMongofs(t *testing.T) {
	fs := New("localhost", "media", "fs")
	fileHandler := http.FileServer(fs)
	server := httptest.NewServer(fileHandler)
	defer server.Close()
	buffer := bytes.NewBuffer([]byte(generateText()))
	name, err := fs.Save("txt", nil, buffer)
	assert.Nil(t, err)
	fileUrl := fmt.Sprintf("%s/%s", server.URL, name)
	resp, err := http.Get(fileUrl)
	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	req, err := http.NewRequest("GET", fileUrl, nil)
	assert.Nil(t, err)
	req.Header.Add("Range", "bytes=14-18,20-23,90-")
	client := http.Client{}
	resp, err = client.Do(req)
	assert.Nil(t, err)
	assert.Equal(t, http.StatusPartialContent, resp.StatusCode)
	mediaType, params, err := mime.ParseMediaType(resp.Header.Get("Content-Type"))
	assert.Nil(t, err)
	assert.True(t, strings.HasPrefix(mediaType, "multipart/"))
	mr := multipart.NewReader(resp.Body, params["boundary"])
	p, err := mr.NextPart()
	assert.Nil(t, err)
	assert.Equal(t, p.Header.Get("Content-Range"), "bytes 14-18/595")
	p, err = mr.NextPart()
	assert.Nil(t, err)
	assert.Equal(t, p.Header.Get("Content-Range"), "bytes 20-23/595")
	p, err = mr.NextPart()
	assert.Nil(t, err)
	assert.Equal(t, p.Header.Get("Content-Range"), "bytes 90-594/595")
	p, err = mr.NextPart()
	assert.NotNil(t, err)
}

func generateText() string {
	return `
	Lorem ipsum dolor sit amet, consetetur sadipscing elitr, sed diam nonumy eirmod tempor invidunt ut labore et dolore magna aliquyam erat, sed diam voluptua. At vero eos et accusam et justo duo dolores et ea rebum. Stet clita kasd gubergren, no sea takimata sanctus est Lorem ipsum dolor sit amet. Lorem ipsum dolor sit amet, consetetur sadipscing elitr, sed diam nonumy eirmod tempor invidunt ut labore et dolore magna aliquyam erat, sed diam voluptua. At vero eos et accusam et justo duo dolores et ea rebum. Stet clita kasd gubergren, no sea takimata sanctus est Lorem ipsum dolor sit amet.
	`
}
