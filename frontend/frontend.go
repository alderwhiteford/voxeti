package frontend

import (
	"io/fs"
	"log"
	"net/url"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

var (
	embedFrontend = false // enable at build time with: "go build -tags embed"
	distDirFS     fs.FS
	srcDirFS      fs.FS
	distIndexHTML fs.FS
)

func RegisterFrontendHandlers(e *echo.Echo) {
	if embedFrontend {
		e.FileFS("/", "index.html", distIndexHTML)
		e.StaticFS("/", distDirFS)
		e.StaticFS("src/", srcDirFS)
		return
	}
	setupProxy(e)
}

func setupProxy(e *echo.Echo) {
	url, err := url.Parse("http://localhost:4000")
	if err != nil {
		log.Fatal(err)
	}

	balancer := middleware.NewRoundRobinBalancer([]*middleware.ProxyTarget{
		{
			URL: url,
		},
	})
	e.Use(middleware.ProxyWithConfig(middleware.ProxyConfig{
		Balancer: balancer,
		Skipper: func(c echo.Context) bool {
			return len(c.Path()) >= 4 && c.Path()[:4] == "/api"
		},
	}))
}
