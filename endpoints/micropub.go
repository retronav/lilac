package endpoints

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"go.karawale.in/lilac/microformats"
	"go.karawale.in/lilac/middleware"
	"go.karawale.in/lilac/post"
	storepkg "go.karawale.in/lilac/store"
	"golang.org/x/exp/slices"
)

func HandleMicropubQuery(persistence storepkg.Persistence) func(*gin.Context) {
	return func(ctx *gin.Context) {

		query := middleware.ArrayQueryParams(ctx.Request.URL.Query())
		switch query.Get("q") {
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
			postPropertiesMap := post.PostToJf2(postProperties)

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
