package main

import (
	"flag"
	"log"
	"mime"
	"net/http"
	s "strings"

	"github.com/InteractiveLecture/media-service/backend"
	"github.com/InteractiveLecture/media-service/mongofs"
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
		id, err := fileBackend.Save(contentType, nil, uploadedFile)
		if err != nil {
			log.Printf("problem saving file %v", err)
			w.WriteHeader(http.StatusInternalServerError)
		}
		w.Header().Set("Location", id)
		w.WriteHeader(http.StatusCreated)
	}
	return http.Handler(http.HandlerFunc(result))
}

func buildUploadHandler(host string) http.Handler {
	backend := mongofs.New(host, "media", "fs")
	uploadHandler := UploadHandler(backend)
	return jwtware.New(
		groupware.New(groupware.DefaultOptions(
			uploadHandler)))
}

func buildDownloadHandler(host string) http.Handler {
	backend := mongofs.New(host, "media", "fs")
	return jwtware.New(http.FileServer(backend))
}

func main() {
	host := flag.String("mongohost", "mongo", "hostname of mongodb")
	r := mux.NewRouter()
	// middleware-chain handlers. Router -> jwt-verification -> group verification-> application
	r.Methods("POST").Path("/").Handler(buildUploadHandler(*host))
	r.Methods("GET").Path("/{id}").Handler(buildDownloadHandler(*host))
	log.Println("listening on 8080")
	// Bind to a port and pass our router in
	http.ListenAndServe(":8080", r)
}
