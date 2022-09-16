package main

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.karawale.in/lilac/endpoints"
	"go.karawale.in/lilac/middleware"
	storepkg "go.karawale.in/lilac/store"
)

func main() {
	store := storepkg.GitStore{
		Path:   "../website",
		Branch: "main",
	}

	store.Init()

	persistence := storepkg.NewPersistence(&store, nil)
	err := persistence.Load()
	if err != nil {
		log.Fatalln(err)
	}

	r := gin.Default()
	r.Use(middleware.BodyParser())

	micropubRouter := r.Group("/micropub")
	micropubRouter.GET("", endpoints.HandleMicropubQuery(persistence))

	http.ListenAndServe(":8080", r)
}
