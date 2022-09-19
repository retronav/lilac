package endpoints

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"path"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/karlseguin/typed"
	"github.com/sirupsen/logrus"
	"github.com/spf13/afero"
	"github.com/spf13/viper"
	"golang.org/x/exp/maps"
	"golang.org/x/exp/slices"
	"karawale.in/go/lilac/microformats"
	"karawale.in/go/lilac/middleware"
	postpkg "karawale.in/go/lilac/post"
	storepkg "karawale.in/go/lilac/store"
	"karawale.in/go/lilac/template"
)

var (
	errorNotFound = fmt.Errorf("not found")
)

func HandleMicropubQuery(persistence storepkg.Persistence) func(*gin.Context) {
	return func(ctx *gin.Context) {
		query := middleware.ArrayQueryParams(ctx.Request.URL.Query())
		switch query.Get("q") {
		case "config":
			serverUrl := ctx.GetString("serverUrl")
			mediaEndpoint, _ := url.JoinPath(serverUrl, "media")
			ctx.JSON(http.StatusOK, gin.H{
				"media-endpoint": mediaEndpoint,
				"syndicate-to":   []string{},
			})
			return
		case "source":
			postUrl := ctx.Query("url")
			if postUrl == "" {
				ctx.Status(http.StatusBadRequest)
			}
			postProperties, exists := persistence.PostProperties.Content[postUrl]
			if !exists {
				ctx.Status(http.StatusNotFound)
				return
			}
			postPropertiesMap := postpkg.PostToJf2(postProperties)

			filterProperties := query["properties"]
			if len(filterProperties) > 0 {

				for key := range postPropertiesMap {
					// "type" key is required for converting to mf2
					if !slices.Contains(append(filterProperties, "type"), key) {
						delete(postPropertiesMap, key)
					}
				}
			}

			mf2, err := microformats.Jf2ToMf2(postPropertiesMap)
			if err != nil {
				ctx.Status(http.StatusInternalServerError)
				return
			}
			if len(filterProperties) > 0 {
				// send only properties if filters are applied
				ctx.JSON(http.StatusOK, gin.H{"properties": mf2.Properties})
			} else {
				ctx.JSON(http.StatusOK, mf2)
			}
		}
	}
}

func HandleMicropubPOST(
	store storepkg.GitStore,
	persistence storepkg.Persistence) func(*gin.Context) {
	return func(ctx *gin.Context) {
		auth, exists := ctx.Get("auth")
		if !exists {
			logrus.Error("auth not present in context?")
			ctx.Status(http.StatusInternalServerError)
			return
		}
		scope := auth.(middleware.IndieauthResponse).Scope

		body, exists := ctx.Get("body")
		if !exists {
			ctx.Status(http.StatusBadRequest)
			return
		}
		bodyMap := body.(map[string]interface{})
		mapKeys := maps.Keys(bodyMap)
		var err error // defined because := creates new variable in switch
		switch {
		case slices.Contains(mapKeys, "type") || slices.Contains(mapKeys, "h"):
			if !scope.Has("create") {
				logrus.Error("insufficient scope to create post")
				ctx.Status(http.StatusForbidden)
				return
			}
			logrus.Info("request wants to create a post")
			var jf2 microformats.Jf2
			switch ctx.ContentType() {
			case gin.MIMEJSON:
				jf2, err = microformats.JsonToJf2(typed.New(bodyMap))
				if err != nil {
					logrus.Error(err)
					ctx.Status(http.StatusBadRequest)
					return
				}
			case gin.MIMEPOSTForm, gin.MIMEMultipartPOSTForm:
				jf2, err = microformats.FormEncodedToJf2(
					castMap[[]string](bodyMap))
				if err != nil {
					logrus.Error(err)
					ctx.Status(http.StatusBadRequest)
					return
				}
			}
			logrus.Infof("parsed jf2: %+v", jf2)
			postUrl, err := createPost(jf2, ctx, store, persistence)
			if err != nil {
				logrus.Error(err)
				ctx.Status(http.StatusInternalServerError)
				return
			}
			ctx.Header("Location", postUrl)
		case slices.Contains(mapKeys, "action") &&
			slices.Contains(mapKeys, "url"):
			url := bodyMap["url"].(string)
			action := bodyMap["action"].(string)

			switch action {
			case "update":
				if !scope.Has("update") {
					logrus.Error("insufficient scope to update post")
					ctx.Status(http.StatusForbidden)
					return
				}
				spec := map[string]map[string]interface{}{}
				for _, specType := range []string{"add", "replace", "delete"} {
					if slices.Contains(mapKeys, specType) {
						spec[specType] = bodyMap[specType].(map[string]interface{})
					}
				}
				logrus.Infof("request wants to update %s", url)
				logrus.Infof("request update spec: %+v", spec)
				if err := updatePost(url, spec, store, persistence); err != nil {
					logrus.Error(err)
					ctx.Status(http.StatusInternalServerError)
					return
				}
			case "delete":
				if !scope.Has("delete") {
					logrus.Error("insufficient scope to delete post")
					ctx.Status(http.StatusForbidden)
					return
				}
				logrus.Infof("request wants to delete %s", url)
				if err := deletePost(url, store, persistence); err != nil {
					logrus.Error(err)
					ctx.Status(http.StatusInternalServerError)
					return
				}
			}
		default:
			logrus.Error("nothing to do")
			ctx.AbortWithStatus(http.StatusBadRequest)
			return
		}
	}
}

func createPost(
	jf2Post map[string]interface{},
	ctx *gin.Context, store storepkg.GitStore,
	persistence storepkg.Persistence) (string, error) {
	postUrl := ""
	post, err := postpkg.Jf2ToPost(jf2Post)
	if err != nil {
		return postUrl, err
	}

	config, exists := ctx.Get("config")
	if !exists {
		return postUrl, fmt.Errorf("no config")
	}
	viper := config.(*viper.Viper)
	tz, err := time.LoadLocation(viper.GetString("micropub.post.timezone"))
	if err != nil {
		return postUrl, err
	}

	postDirs := viper.Get("micropub.post.dir").(map[string]interface{})
	postUrlPrefixes := viper.Get("micropub.post.url").(map[string]interface{})

	postDir, exists := postDirs[string(post.POST_TYPE)].(string)
	if !exists {
		return postUrl, fmt.Errorf("post %s not configured", post.POST_TYPE)
	}
	postUrlPrefix, exists := postUrlPrefixes[string(post.POST_TYPE)].(string)
	if !exists {
		return postUrl, fmt.Errorf("post %s not configured", post.POST_TYPE)
	}

	postTimeDir := time.Now().In(tz).Format("2006/01/02")

	postFullDir := path.Join(store.Path, postDir, postTimeDir)
	if _, err := os.Stat(postFullDir); errors.Is(err, os.ErrNotExist) {
		if err := os.MkdirAll(postFullDir, 0777); err != nil {
			return postUrl, err
		}
	}
	fileBasenameInt, err := fileCount(postFullDir, store.Fs)
	if err != nil {
		return postUrl, err
	}
	fileBasename := fmt.Sprintf("%02d", fileBasenameInt+1)

	filename := path.Join(postDir, postTimeDir, fileBasename+".md")

	relMe := viper.GetString("micropub.me")
	postUrl, err = url.JoinPath(relMe, postUrlPrefix, postTimeDir, fileBasename)
	if err != nil {
		return postUrl, err
	}

	postTmpl := viper.GetString("micropub.post.template")
	rendered, err := template.RenderTemplate(post, postTmpl)
	if err != nil {
		return postUrl, err
	}
	if err = os.WriteFile(path.Join(store.Path, filename), []byte(rendered), 0666); err != nil {
		return postUrl, err
	}
	logrus.Infof("post saved at: %s", filename)

	persistence.PostMappings.Content[postUrl] = filename
	persistence.PostProperties.Content[postUrl] = post

	if err = persistence.Dump(); err != nil {
		return postUrl, err
	}
	if err = store.Sync(":robot: Updates from Lilac"); err != nil {
		return postUrl, err
	}

	return postUrl, nil
}

func updatePost(
	url string,
	spec map[string]map[string]interface{},
	store storepkg.GitStore,
	persistence storepkg.Persistence) error {
	post, exists := persistence.PostProperties.Content[url]
	if !exists {
		return errorNotFound
	}
	postLocation := persistence.PostMappings.Content[url]
	jf2 := postpkg.PostToJf2(post)

	if add, exists := spec["add"]; exists {
		jf2.Add(add)
	}
	if replace, exists := spec["replace"]; exists {
		jf2.Replace(replace)
	}
	if toDelete, exists := spec["delete"]; exists {
		jf2.Delete(toDelete)
	}

	post, err := postpkg.Jf2ToPost(jf2)
	if err != nil {
		return err
	}

	updatedTime := time.Now()
	post.Updated = &updatedTime

	postTmpl := viper.GetString("micropub.post.template")
	rendered, err := template.RenderTemplate(post, postTmpl)
	if err != nil {
		return err
	}

	file, err := store.Fs.OpenFile(postLocation, os.O_RDWR, 0666)
	if err != nil {
		return err
	}
	if err = file.Truncate(0); err != nil {
		return err
	}
	if _, err = file.WriteString(rendered); err != nil {
		return err
	}
	persistence.PostProperties.Content[url] = post
	if err = persistence.Dump(); err != nil {
		return err
	}
	if err = store.Sync(":robot: Updates from Lilac"); err != nil {
		return err
	}

	return nil
}

func deletePost(
	url string,
	store storepkg.GitStore,
	persistence storepkg.Persistence) error {
	mappingKeys := maps.Keys(persistence.PostMappings.Content)
	propertiesKeys := maps.Keys(persistence.PostMappings.Content)

	if slices.Contains(mappingKeys, url) &&
		slices.Contains(propertiesKeys, url) {
		pathOnStore := persistence.PostMappings.Content[url]
		if err := store.Fs.Remove(pathOnStore); err != nil {
			return err
		}
		if _, err := store.Fs.Create(pathOnStore + ".deleted"); err != nil {
			return err
		}

		delete(persistence.PostMappings.Content, url)
		delete(persistence.PostProperties.Content, url)

		if err := persistence.Dump(); err != nil {
			return err
		}
		if err := store.Sync(":robot: Updates from Lilac"); err != nil {
			return err
		}
	} else {
		return errorNotFound
	}

	return nil
}

func fileCount(path string, fs afero.Fs) (int, error) {
	n := 0
	files, err := os.ReadDir(path)
	if err != nil {
		return n, err
	}
	for _, file := range files {
		if !file.IsDir() {
			n++
		}
	}
	return n, nil
}

func castMap[T any](m map[string]interface{}) map[string]T {
	to := make(map[string]T)
	for key, value := range m {
		to[key] = value.(T)
	}
	return to
}
