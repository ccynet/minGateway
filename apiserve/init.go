package apiserve

import (
	"fmt"
	"minGateway/apiserve/router"
	"net/http"
	"time"

	"github.com/gorilla/mux"
)

func Run() *http.Server {

	muxRouter := mux.NewRouter()
	router.InitRouter(muxRouter)

	srv := &http.Server{
		Handler:      muxRouter,
		Addr:         ":9900",
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			panic(err.Error())
		}
	}()

	fmt.Println("\nAPI监听端口:9900")
	return srv
}
