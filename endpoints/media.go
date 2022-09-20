package endpoints

import (
	"errors"
	"io"
	"io/fs"
	"net/http"
	"net/url"
	"path"
	"time"

	"github.com/davidbyttow/govips/v2/vips"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
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

		vips.Startup(nil)
		defer vips.Shutdown()

		img, err := vips.LoadImageFromBuffer(fileBlob, vips.NewImportParams())
		if err != nil {
			logrus.Error(err)
			ctx.Status(http.StatusInternalServerError)
			return
		}

		exportParams := vips.NewWebpExportParams()
		exportParams.StripMetadata = true

		processedBlob, _, err := img.ExportWebp(exportParams)
		if err != nil {
			logrus.Error(err)
			ctx.Status(http.StatusInternalServerError)
			return
		}

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

		mediaUrlDir := viper.GetString("micropub.media.url")
		meUrl := viper.GetString("micropub.me")
		mediaUrl, _ := url.JoinPath(meUrl, mediaUrlDir, filename)

		ctx.Header("Location", mediaUrl)
		ctx.Status(http.StatusCreated)
	}
}
