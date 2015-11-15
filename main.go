package main

import (
	//"github.com/gorilla/context"
	"github.com/InteractiveLecture/middlewares/authware"
	"github.com/InteractiveLecture/middlewares/groupware"
	"github.com/InteractiveLecture/middlewares/jwtware"
	"github.com/InteractiveLecture/serviceclient"
	"github.com/InteractiveLecture/serviceclient/cacheadapter"
	"github.com/gorilla/mux"
	"log"
	"net/http"
	"time"
)

func main() {
	r := mux.NewRouter()
	adapter := cacheadapter.New("discovery:8500", 10*time.Second, 5*time.Second, 3)
	serviceclient.Configure(adapter, "acl-service", "authentication-service")
	// middleware-chain handlers. Router -> jwt-verification -> application
	r.Methods("POST").
		Path("/").
		Handler(jwtware.New(groupware.New(UploadHandler(), "officer", "assistant")))
	r.Methods("GET").
		Path("/{id}").
		Handler(jwtware.New(authware.New(DownloadHandler(), "media", "read")))
	log.Println("listening on 8000")
	// Bind to a port and pass our router in
	http.ListenAndServe(":8000", r)
}
