package middleware

import (
	"encoding/json"
	"mime"
	"net/http"
	"net/url"
	"strings"

	"github.com/docker/go-units"
	"github.com/gin-gonic/gin"
)

// type ContextKey string

// func BodyParser(next http.Handler) http.Handler {
// 	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
// 		ct := r.Header.Get("content-type")
// 		switch {
// 		case strings.HasPrefix(ct, "application/json"):
// 			{
// 				body := new(bytes.Buffer)
// 				_, err := io.Copy(body, r.Body)
// 				if err != nil {
// 					return
// 				}
// 				defer r.Body.Close()
// 				var parsed map[string]interface{}
// 				err = json.Unmarshal(body.Bytes(), &parsed)
// 				if err != nil {
// 					return
// 				}
// 				ctx := context.WithValue(r.Context(), ContextKey("body"), parsed)
// 				next.ServeHTTP(w, r.WithContext(ctx))
// 			}
// 		case strings.HasPrefix(ct, "multipart/form-data"):
// 			{
// 				err := r.ParseMultipartForm(1000)
// 				if err != nil {
// 					return
// 				}
// 				parsed := ArrayQueryParams(r.MultipartForm.Value)

// 				ctx := context.WithValue(r.Context(), ContextKey("body"), parsed)
// 				next.ServeHTTP(w, r.WithContext(ctx))
// 			}
// 		case strings.HasPrefix(ct, "application/x-www-form-urlencoded"):
// 			{
// 				err := r.ParseForm()
// 				if err != nil {
// 					return
// 				}
// 				parsed := ArrayQueryParams(r.PostForm)
// 				ctx := context.WithValue(r.Context(), ContextKey("body"), parsed)
// 				next.ServeHTTP(w, r.WithContext(ctx))
// 			}
// 		default:
// 			{
// 				w.WriteHeader(http.StatusUnprocessableEntity)
// 			}
// 			next.ServeHTTP(w, r)
// 		}
// 	})
// }

func BodyParser() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		ctHeader := ctx.GetHeader("content-type")
		if ctHeader == "" {
			return
		}
		ct, _, err := mime.ParseMediaType(ctHeader)
		if err != nil {
			panic("unknown content type")
		}
		switch ct {
		case gin.MIMEJSON:
			var body map[string]interface{}
			decoder := json.NewDecoder(ctx.Request.Body)
			err = decoder.Decode(&body)
			if err != nil {
				ctx.Status(http.StatusBadRequest)
			}
			ctx.Set("body", body)
		case gin.MIMEPOSTForm:
			err = ctx.Request.ParseForm()
			if err != nil {
				ctx.Status(http.StatusBadRequest)
			}
			ctx.Set("body", ArrayQueryParams(ctx.Request.PostForm))
		case gin.MIMEMultipartPOSTForm:
			err = ctx.Request.ParseMultipartForm(10 * units.MB)
			if err != nil {
				ctx.Status(http.StatusBadRequest)
			}
			ctx.Set("body", ArrayQueryParams(ctx.Request.MultipartForm.Value))
		}
	}
}

func ArrayQueryParams(params url.Values) url.Values {
	for key, value := range params {
		if strings.HasSuffix(key, "[]") {
			normalizedKey := strings.TrimRight(key, "[]")
			delete(params, key)
			params[normalizedKey] = value
		}
	}
	return params
}
