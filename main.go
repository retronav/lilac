package main

import (
	"fmt"
	"log"
	"net"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"karawale.in/go/lilac/endpoints"
	"karawale.in/go/lilac/middleware"
	storepkg "karawale.in/go/lilac/store"
	"karawale.in/go/lilac/version"
)

func main() {
	logrus.Infof("Lilac version %s", version.Version)
	if version.Version != "dev" {
		gin.SetMode(gin.ReleaseMode)
	}
	viper.AddConfigPath(".")
	viper.SetConfigName("lilac")
	viper.SetConfigType("toml")
	err := viper.ReadInConfig()
	if err != nil {
		panic(fmt.Errorf("fatal error config file: %w", err))
	}
	store := storepkg.GitStore{
		Path:   viper.GetString("store.git.path"),
		Branch: viper.GetString("store.git.branch"),
	}

	if err = store.Init(); err != nil {
		log.Fatalf("error loading store: %s", err)
	}

	persistence := storepkg.NewPersistence(&store, "")
	if err = persistence.Load(); err != nil {
		log.Fatalf("error loading persistence: %s", err)
	}

	r := gin.Default()
	r.Use(middleware.BodyParser())
	r.Use(func(ctx *gin.Context) {
		ctx.Set("config", viper.GetViper())
	})

	micropubRouter := r.Group("/micropub")
	micropubRouter.Use(
		middleware.Indieauth(viper.GetString("micropub.me"),
			viper.GetString("micropub.token_endpoint")))
	micropubRouter.GET("", endpoints.HandleMicropubQuery(persistence))
	micropubRouter.POST("", endpoints.HandleMicropubPOST(store, persistence))

	mediaRouter := r.Group("/media")
	mediaRouter.Use(
		middleware.Indieauth(viper.GetString("micropub.me"),
			viper.GetString("micropub.token_endpoint")))
	mediaRouter.POST("", endpoints.HandleMediaUpload(store))

	serveOn := viper.GetString("serve_on")
	switch serveOn {
	case "port":
		addr := fmt.Sprintf(":%d", viper.GetInt("port"))
		if err = http.ListenAndServe(addr, r); err != nil {
			logrus.Fatal(err)
		}
	case "unixsock":
		listener, err := net.Listen("unix", viper.GetString("unix_socket"))
		if err != nil {
			logrus.Fatal(err)
		}
		if err = http.Serve(listener, r); err != nil {
			logrus.Fatal(err)
		}
	}
}
