package swaggerui

import (
	"embed"
	"io/fs"
	"net/http"
	"path/filepath"
	"regexp"

	htmlTemplate "html/template"
	textTemplate "text/template"

	gin "github.com/gin-gonic/gin"
	httpheaders "github.com/go-http-utils/headers"
)

//go:embed files/*
var files embed.FS

// FS holds embedded swagger ui files
var FS, _ = fs.Sub(files, "files")

func Add(engine *gin.Engine, relativePath string, specPath string) {
	group := engine.Group(relativePath)
	handler := handle(relativePath, specPath)
	group.GET("/*filepath", handler)
	group.HEAD("/*filepath", handler)
}

func handle(relativePath string, specPath string) gin.HandlerFunc {
	var config = swaggerConfig{
		URL:                      specPath,
		DeepLinking:              true,
		DocExpansion:             "list",
		DefaultModelsExpandDepth: 1,
		Oauth2RedirectURL: "`${window.location.protocol}//${window.location.host}$" +
			"{window.location.pathname.split('/').slice(0, window.location.pathname.split('/').length - 1).join('/')}" +
			"/oauth2-redirect.html`",
		Title:                 "Swagger UI",
		PersistAuthorization:  false,
		Oauth2DefaultClientID: "",
		Oauth2UsePkce:         false,
	}
	var fs http.FileSystem = http.FS(FS)
	fileServer := http.StripPrefix(relativePath, http.FileServer(fs))
	// create a template with name
	index, _ := htmlTemplate.New("swagger_index.html").Parse(swaggerIndexTpl)
	js, _ := textTemplate.New("swagger_index.js").Parse(swaggerJSTpl)
	css, _ := textTemplate.New("swagger_index.css").Parse(swaggerStyleTpl)
	var matcher = regexp.MustCompile(`(.*)(index\.html|index\.css|swagger-initializer\.js|favicon-16x16\.png|favicon-32x32\.png|/oauth2-redirect\.html|swagger-ui\.css|swagger-ui\.css\.map|swagger-ui\.js|swagger-ui\.js\.map|swagger-ui-bundle\.js|swagger-ui-bundle\.js\.map|swagger-ui-standalone-preset\.js|swagger-ui-standalone-preset\.js\.map)[?|.]*`)
	return func(ctx *gin.Context) {
		if ctx.Request.Method != http.MethodGet {
			ctx.AbortWithStatus(http.StatusMethodNotAllowed)
			return
		}
		file := ctx.Param("filepath")
		if file == "/" {
			file = "index.html"
		}
		matches := matcher.FindStringSubmatch(file)
		if len(matches) != 3 {
			ctx.String(http.StatusNotFound, http.StatusText(http.StatusNotFound))
			return
		}
		path := matches[2]
		switch filepath.Ext(path) {
		case ".html":
			ctx.Header(httpheaders.ContentType, contentTypeHTML)
		case ".css":
			ctx.Header(httpheaders.ContentType, contentTypeTextCSS)
		case ".js":
			ctx.Header(httpheaders.ContentType, contentTypeJavascript)
		case ".png":
			ctx.Header(httpheaders.ContentType, contentTypeImagePng)
		case ".json":
			ctx.Header(httpheaders.ContentType, contentTypeJSON)
		}
		switch path {
		case "index.html":
			_ = index.Execute(ctx.Writer, config)
		case "index.css":
			_ = css.Execute(ctx.Writer, config)
		case "swagger-initializer.js":
			_ = js.Execute(ctx.Writer, config)
		default:
			// Check if file exists and/or if we have permission to access it
			f, err := fs.Open(file)
			if err != nil {
				ctx.Writer.WriteHeader(http.StatusNotFound)
				return
			}
			f.Close()
			fileServer.ServeHTTP(ctx.Writer, ctx.Request)
		}
	}
}

const (
	contentTypeJSON       = "application/json; charset=utf-8"
	contentTypeJavascript = "application/javascript"
	contentTypeHTML       = "text/html; charset=utf-8"
	contentTypeText       = "text/plain; charset=utf-8"
	contentTypeTextCSS    = "text/css; charset=utf-8"
	contentTypeImagePng   = "image/png"
)

type swaggerConfig struct {
	// The url pointing to API definition (normally swagger.json or swagger.yaml). Default is `doc.json`.
	URL                      string
	DocExpansion             string
	Title                    string
	Oauth2RedirectURL        htmlTemplate.JS
	DefaultModelsExpandDepth int
	DeepLinking              bool
	PersistAuthorization     bool
	Oauth2DefaultClientID    string
	Oauth2UsePkce            bool
}

const swaggerStyleTpl = `
html
{
    box-sizing: border-box;
    overflow: -moz-scrollbars-vertical;
    overflow-y: scroll;
}
*,
*:before,
*:after
{
    box-sizing: inherit;
}

body {
  margin:0;
  background: #fafafa;
}
`

const swaggerJSTpl = `
window.onload = function() {
  // Build a system
  const ui = SwaggerUIBundle({
    url: "{{.URL}}",
    dom_id: '#swagger-ui',
    validatorUrl: null,
    oauth2RedirectUrl: {{.Oauth2RedirectURL}},
    persistAuthorization: {{.PersistAuthorization}},
    presets: [
      SwaggerUIBundle.presets.apis,
      SwaggerUIStandalonePreset
    ],
    plugins: [
      SwaggerUIBundle.plugins.DownloadUrl
    ],
	layout: "StandaloneLayout",
    docExpansion: "{{.DocExpansion}}",
	deepLinking: {{.DeepLinking}},
	defaultModelsExpandDepth: {{.DefaultModelsExpandDepth}}
  })

  const defaultClientId = "{{.Oauth2DefaultClientID}}";
  if (defaultClientId) {
    ui.initOAuth({
      clientId: defaultClientId,
      usePkceWithAuthorizationCodeGrant: {{.Oauth2UsePkce}}
    })
  }

  window.ui = ui
}
`

const swaggerIndexTpl = `<!-- HTML for static distribution bundle build -->
<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="UTF-8">
  <title>{{.Title}}</title>
  <link rel="stylesheet" type="text/css" href="./swagger-ui.css" >
  <link rel="icon" type="image/png" href="./favicon-32x32.png" sizes="32x32" />
  <link rel="icon" type="image/png" href="./favicon-16x16.png" sizes="16x16" />
  <link rel="stylesheet" type="text/css" href="index.css" />
</head>

<body>

<svg xmlns="http://www.w3.org/2000/svg" xmlns:xlink="http://www.w3.org/1999/xlink" style="position:absolute;width:0;height:0">
  <defs>
    <symbol viewBox="0 0 20 20" id="unlocked">
          <path d="M15.8 8H14V5.6C14 2.703 12.665 1 10 1 7.334 1 6 2.703 6 5.6V6h2v-.801C8 3.754 8.797 3 10 3c1.203 0 2 .754 2 2.199V8H4c-.553 0-1 .646-1 1.199V17c0 .549.428 1.139.951 1.307l1.197.387C5.672 18.861 6.55 19 7.1 19h5.8c.549 0 1.428-.139 1.951-.307l1.196-.387c.524-.167.953-.757.953-1.306V9.199C17 8.646 16.352 8 15.8 8z"></path>
    </symbol>

    <symbol viewBox="0 0 20 20" id="locked">
      <path d="M15.8 8H14V5.6C14 2.703 12.665 1 10 1 7.334 1 6 2.703 6 5.6V8H4c-.553 0-1 .646-1 1.199V17c0 .549.428 1.139.951 1.307l1.197.387C5.672 18.861 6.55 19 7.1 19h5.8c.549 0 1.428-.139 1.951-.307l1.196-.387c.524-.167.953-.757.953-1.306V9.199C17 8.646 16.352 8 15.8 8zM12 8H8V5.199C8 3.754 8.797 3 10 3c1.203 0 2 .754 2 2.199V8z"/>
    </symbol>

    <symbol viewBox="0 0 20 20" id="close">
      <path d="M14.348 14.849c-.469.469-1.229.469-1.697 0L10 11.819l-2.651 3.029c-.469.469-1.229.469-1.697 0-.469-.469-.469-1.229 0-1.697l2.758-3.15-2.759-3.152c-.469-.469-.469-1.228 0-1.697.469-.469 1.228-.469 1.697 0L10 8.183l2.651-3.031c.469-.469 1.228-.469 1.697 0 .469.469.469 1.229 0 1.697l-2.758 3.152 2.758 3.15c.469.469.469 1.229 0 1.698z"/>
    </symbol>

    <symbol viewBox="0 0 20 20" id="large-arrow">
      <path d="M13.25 10L6.109 2.58c-.268-.27-.268-.707 0-.979.268-.27.701-.27.969 0l7.83 7.908c.268.271.268.709 0 .979l-7.83 7.908c-.268.271-.701.27-.969 0-.268-.269-.268-.707 0-.979L13.25 10z"/>
    </symbol>

    <symbol viewBox="0 0 20 20" id="large-arrow-down">
      <path d="M17.418 6.109c.272-.268.709-.268.979 0s.271.701 0 .969l-7.908 7.83c-.27.268-.707.268-.979 0l-7.908-7.83c-.27-.268-.27-.701 0-.969.271-.268.709-.268.979 0L10 13.25l7.418-7.141z"/>
    </symbol>


    <symbol viewBox="0 0 24 24" id="jump-to">
      <path d="M19 7v4H5.83l3.58-3.59L8 6l-6 6 6 6 1.41-1.41L5.83 13H21V7z"/>
    </symbol>

    <symbol viewBox="0 0 24 24" id="expand">
      <path d="M10 18h4v-2h-4v2zM3 6v2h18V6H3zm3 7h12v-2H6v2z"/>
    </symbol>

  </defs>
</svg>

<div id="swagger-ui"></div>

<script src="./swagger-ui-bundle.js"> </script>
<script src="./swagger-ui-standalone-preset.js"> </script>
<script src="./swagger-initializer.js"> </script>
</body>

</html>
`
