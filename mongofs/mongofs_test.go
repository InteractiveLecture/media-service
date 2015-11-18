package mongofs

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMongofs(t *testing.T) {
	fs := New("localhost", "media", "fs")
	fileHandler := http.FileServer(fs)
	server := httptest.NewServer(fileHandler)
	file, err := os.Open("mongofs.go")
	assert.Nil(t, err)
	name, err := fs.Save("txt", file)
	assert.Nil(t, err)

}
