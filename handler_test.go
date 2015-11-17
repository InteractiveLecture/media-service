package main

import (
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/InteractiveLecture/media-service/backend"
	"github.com/InteractiveLecture/media-service/backend/mocks"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func mockExtractor(r *http.Request) (string, error) {
	return "123", nil
}

func TestDownloadHandler(t *testing.T) {
	controller := gomock.NewController(t)
	defer controller.Finish()
	backendMock := filebackendmocks.NewMockFileBackend(controller)
	f, err := os.Open("mime.types")
	defer f.Close()
	assert.Nil(t, err)
	stat, _ := f.Stat()
	size := stat.Size()
	log.Printf("size is %d", size)
	meta := backend.NewMeta("mime.types", "text/plain", size, nil)
	backendMock.EXPECT().LoadById("123").Return(f, meta, nil)
	handler := DownloadHandler(backendMock, mockExtractor)
	mediaServer := httptest.NewServer(handler)
	defer mediaServer.Close()
	client := &http.Client{}
	req, err := http.NewRequest("GET", mediaServer.URL, nil)
	assert.Nil(t, err)
	req.Header.Add("Range", "bytes=-3,10-20,25-")
	resp, err := client.Do(req)
	assert.Nil(t, err)
	defer resp.Body.Close()
	log.Println(resp.Header)
	log.Println(resp.StatusCode)
}
