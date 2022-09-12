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
	r.Use(func(ctx *gin.Context) {
		ctx.Set("store", store)
		ctx.Set("persistence", persistence)
	})

	r.GET("/micropub", endpoints.HandleMicropubQuery)

	http.ListenAndServe(":8080", r)
}
