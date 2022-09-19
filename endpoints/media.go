package endpoints

import (
	"errors"
	"io"
	"io/fs"
	"net/http"
	"path"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"gopkg.in/gographics/imagick.v2/imagick"
	"karawale.in/go/lilac/middleware"
	storepkg "karawale.in/go/lilac/store"
)

func HandleMediaUpload(store storepkg.GitStore) func(*gin.Context) {
	return func(ctx *gin.Context) {
		auth, exists := ctx.Get("auth")
		if !exists {
			logrus.Error("auth not present in context?")
			ctx.Status(http.StatusInternalServerError)
			return
		}
		scope := auth.(middleware.IndieauthResponse).Scope

		if !scope.Has("media") {
			logrus.Error("insufficient scope for media upload")
			ctx.Status(http.StatusForbidden)
			return
		}

		media, err := ctx.FormFile("file")
		if err != nil {
			logrus.Error(err)
			ctx.Status(http.StatusBadRequest)
			return
		}
		config, exists := ctx.Get("config")
		if !exists {
			logrus.Error("no config")
			ctx.Status(http.StatusInternalServerError)
			return
		}
		viper := config.(*viper.Viper)

		timestampFormat := "20060102150405" // yyyyMMddhhmmss
		mediaDir := viper.GetString("micropub.media.dir")

		if _, err = store.Fs.Stat(mediaDir); err != nil {
			if errors.Is(err, fs.ErrNotExist) {
				if err = store.Fs.MkdirAll(mediaDir, 0777); err != nil {
					logrus.Error(err)
					ctx.Status(http.StatusInternalServerError)
					return
				}
			} else {
				logrus.Error(err)
				ctx.Status(http.StatusInternalServerError)
				return
			}
		}

		fileToRead, err := media.Open()
		if err != nil {
			logrus.Error(err)
			ctx.Status(http.StatusInternalServerError)
			return
		}
		fileBlob, err := io.ReadAll(fileToRead)
		if err != nil {
			logrus.Error(err)
			ctx.Status(http.StatusInternalServerError)
			return
		}

		imagick.Initialize()
		defer imagick.Terminate()

		mw := imagick.NewMagickWand()
		defer mw.Destroy()

		if err = mw.ReadImageBlob(fileBlob); err != nil {
			logrus.Error(err)
			ctx.Status(http.StatusInternalServerError)
			return
		}

		if err = mw.SetImageFormat("webp"); err != nil {
			logrus.Error(err)
			ctx.Status(http.StatusInternalServerError)
			return
		}
		if err = mw.SetImageCompressionQuality(80); err != nil {
			logrus.Error(err)
			ctx.Status(http.StatusInternalServerError)
			return
		}

		processedBlob := mw.GetImageBlob()

		filename := time.Now().Format(timestampFormat) + ".webp"

		file, err := store.Fs.Create(path.Join(mediaDir, filename))
		if err != nil {
			logrus.Error(err)
			ctx.Status(http.StatusInternalServerError)
			return
		}
		if _, err = file.Write(processedBlob); err != nil {
			logrus.Error(err)
			ctx.Status(http.StatusInternalServerError)
			return
		}

		if err = store.Sync(":robot: Updates from Lilac"); err != nil {
			logrus.Error(err)
			ctx.Status(http.StatusInternalServerError)
			return
		}

		ctx.Status(http.StatusCreated)
	}
}
