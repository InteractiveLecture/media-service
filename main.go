package main

import (
	"fmt"
	//"github.com/gorilla/context"
	"github.com/gorilla/mux"
	//	"github.com/richterrettich/media-service/requestutils"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"io"
	"log"
	"mime"
	"net/http"
	"strconv"
	s "strings"
)

func getSession() (session *mgo.Session, err error) {
	return mgo.Dial("mongo")
}

func uploadHandlerFunc(w http.ResponseWriter, r *http.Request) {
	log.Println("Uploading file...")
	uploadedFile, header, err := r.FormFile("data")
	if err != nil {
		log.Printf("problem with file %v", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	defer uploadedFile.Close()
	log.Printf("got file %s", header.Filename)
	parts := s.Split(header.Filename, ".")
	extension := parts[len(parts)-1]
	contentType := mime.TypeByExtension("." + extension)
	log.Printf("determined %s as content type", contentType)
	session, connectionError := getSession()
	if connectionError != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	defer session.Close()
	//Mongodb id will be the name of the file. See https://godoc.org/labix.org/v2/mgo#GridFS.Create
	file, err := session.DB("media-service").GridFS("fs").Create("")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	file.SetContentType(contentType)
	defer file.Close()
	_, copyErr := io.Copy(file, uploadedFile)
	if copyErr != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	id := file.Id().(bson.ObjectId).Hex()
	log.Printf("id is %s", id)
	w.Header().Set("Location", id)
	w.WriteHeader(http.StatusCreated)

}

func downloadHandlerFunc(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := bson.ObjectIdHex(vars["id"])
	session, connectionError := getSession()
	if connectionError != nil {
		w.WriteHeader(http.StatusInternalServerError)
	}
	defer session.Close()
	file, err := session.DB("media-service").GridFS("fs").OpenId(id)
	if err != nil {
		if err == mgo.ErrNotFound {
			log.Println("Not Found error")
			w.WriteHeader(http.StatusNotFound)
			return
		}
		log.Println("Other error")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	defer file.Close()
	log.Printf("Got file with size %d", file.Size())
	rangeRequest := r.Header.Get("Range")
	defer file.Close()
	w.Header().Set("Content-Type", file.ContentType())
	if rangeRequest == "" {
		handleNormalRequest(file, w, r)
	} else {
		handleRangeRequest(file, w, rangeRequest)
	}
}

func handleNormalRequest(file *mgo.GridFile, w http.ResponseWriter, r *http.Request) {
	log.Printf("Downloading file...")
	w.Header().Set("Content-Length", strconv.FormatInt(file.Size(), 10))
	amount, err := io.Copy(w, file)
	if err != nil {
		log.Printf("connection aborted %v", err)
	}
	log.Printf("copied %d to client.", amount)
}
func handleRangeRequest(file *mgo.GridFile, w http.ResponseWriter, rangeRequest string) {
	ranges, err := ParseRangeRequest(rangeRequest)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	for _, r := range ranges {
		from := r.From
		if from < 0 {
			from = file.Size() - r.To
		}
		to := r.To
		if to < 0 {
			to = file.Size() - 1
		}
		file.Seek(from, 0)
		contentLength := (to - from) + 1
		w.Header().Add("Accept-Ranges", "bytes")
		w.Header().Add("Content-Length", strconv.FormatInt(contentLength, 10))
		w.Header().Add("Content-Range", fmt.Sprintf("bytes %d-%d/%d", from, to, file.Size()))
		w.WriteHeader(http.StatusPartialContent)
		amount, err := io.CopyN(w, file, contentLength)
		if err != nil {
			log.Printf("connection aborted %v", err)
		}
		log.Printf("copied %d to client.", amount)
	}
}

func main() {
	r := mux.NewRouter()
	uploadHanlder := http.HandlerFunc(uploadHandlerFunc)
	downloadHandler := http.HandlerFunc(downloadHandlerFunc)
	// middleware-chain handlers. Router -> jwt-verification -> application
	r.Methods("POST").Path("/").Handler(JwtMiddleware(uploadHanlder))
	r.Methods("GET").Path("/{id}").Handler(JwtMiddleware(downloadHandler))
	log.Println("listening on 8000")
	// Bind to a port and pass our router in
	http.ListenAndServe(":8000", r)
}
