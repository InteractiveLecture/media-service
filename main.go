package main

import (
	"errors"
	"fmt"
	"io"
	"log"
	"mime"
	"net/http"
	"strconv"
	s "strings"

	"github.com/InteractiveLecture/media-service/backend"
	"github.com/InteractiveLecture/media-service/mongobackend"
	"github.com/InteractiveLecture/middlewares/groupware"
	"github.com/InteractiveLecture/middlewares/jwtware"
	"github.com/gorilla/mux"
)

func UploadHandler(fileBackend backend.FileBackend) http.Handler {
	result := func(w http.ResponseWriter, r *http.Request) {
		log.Println("Uploading file...")
		uploadedFile, header, err := r.FormFile("data")
		if err != nil {
			log.Printf("problem with file %v", err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		defer uploadedFile.Close()
		parts := s.Split(header.Filename, ".")
		extension := parts[len(parts)-1]
		contentType := mime.TypeByExtension("." + extension)
		log.Printf("determined %s as content type", contentType)
		log.Printf("got file %s", header.Filename)
		id, err := fileBackend.Save("", contentType, nil, uploadedFile)
		if err != nil {
			log.Printf("problem saving file %v", err)
			w.WriteHeader(http.StatusInternalServerError)
		}
		w.Header().Set("Location", id)
		w.WriteHeader(http.StatusCreated)
	}
	return http.Handler(http.HandlerFunc(result))
}

type Extractor func(r *http.Request) (string, error)

func DownloadHandler(fileBackend backend.FileBackend, extractor Extractor) http.Handler {
	result := func(w http.ResponseWriter, r *http.Request) {
		id, err := extractor(r)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		rangeRequest := r.Header.Get("Range")
		file, meta, err := fileBackend.LoadById(id)
		defer file.Close()
		w.Header().Set("Content-Type", meta.ContentType)
		if rangeRequest != "" {
			handleRangeRequest(file, meta, w, rangeRequest)
			return
		}
		handleNormalRequest(file, meta, w)
	}
	return http.Handler(http.HandlerFunc(result))
}

func downloadHandlerFunc(w http.ResponseWriter, r *http.Request) {
}

func handleNormalRequest(file backend.ReadSeekCloser, meta *backend.FileMeta, w http.ResponseWriter) {
	log.Printf("Downloading file...")
	w.Header().Set("Content-Length", strconv.FormatInt(meta.Size, 10))
	amount, err := io.Copy(w, file)
	if err != nil {
		log.Printf("connection aborted %v", err)
	}
	log.Printf("copied %d to client.", amount)
}

func handleRangeRequest(file backend.ReadSeekCloser, meta *backend.FileMeta, w http.ResponseWriter, rangeRequest string) {
	ranges, err := ParseRangeRequest(rangeRequest, meta.Size)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	if len(ranges) > 1 {
		//TODO handle multipart/byteranges response
		for _, r := range ranges {

		}
	} else {
		r := ranges[0]
		file.Seek(r.From, 0)
		w.Header().Add("Accept-Ranges", "bytes")
		w.Header().Add("Content-Length", strconv.FormatInt(r.ContentLength(), 10))
		w.Header().Add("Content-Range", fmt.Sprintf("bytes %d-%d/%d", r.From, r.To, meta.Size))
		w.WriteHeader(http.StatusPartialContent)
		amount, err := io.CopyN(w, file, r.ContentLength())
		if err != nil {
			log.Printf("connection aborted %v", err)
		}
		log.Printf("copied %d to client.", amount)

	}
}

func chunkHeader(contentType, contentRange string, completeSize int64) (string, int64) {
	result := fmt.Sprintf("--3d6b6a416f9b5\nContent-Type: %s\nContent-Range: bytes %s/%d", contentType, contentRange)
}

func buildUploadHandler() http.Handler {
	backend := mongobackend.New("mongo", "media", "fs")
	uploadHandler := UploadHandler(backend)
	return jwtware.New(
		groupware.New(groupware.DefaultOptions(
			uploadHandler)))
}

func buildDownloadHandler() http.Handler {
	backend := mongobackend.New("mongo", "media", "fs")
	extractor := func(r *http.Request) (string, error) {
		vars := mux.Vars(r)
		id, ok := vars["id"]
		if !ok {
			return "", errors.New("could not extract id")
		}
		return id, nil
	}
	handler := DownloadHandler(backend, extractor)

	return jwtware.New(handler)
}

func main() {
	r := mux.NewRouter()
	// middleware-chain handlers. Router -> jwt-verification -> group verification-> application
	r.Methods("POST").Path("/").Handler(buildUploadHandler())
	r.Methods("GET").Path("/{id}").Handler(buildDownloadHandler())
	log.Println("listening on 8000")
	// Bind to a port and pass our router in
	http.ListenAndServe(":8000", r)
}
