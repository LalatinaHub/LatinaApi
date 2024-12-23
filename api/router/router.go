package router

import (
	"fmt"
	"net/http"
	"os"

	getRoute "github.com/LalatinaHub/LatinaApi/api/get"
	"github.com/LalatinaHub/LatinaApi/api/middleware"
	parseRoute "github.com/LalatinaHub/LatinaApi/api/parse"
	subfinderRoute "github.com/LalatinaHub/LatinaApi/api/subfinder"
	"github.com/gin-contrib/static"
	"github.com/gin-gonic/gin"
)

var (
	Router = gin.Default()
	Port   = os.Getenv("PORT")
	Public = "web/public/"
)

func Start() {
	Router.SetTrustedProxies(nil)

	Router.NoRoute(func(c *gin.Context) {
		html, _ := os.ReadFile(fmt.Sprintf("%s404.html", Public))

		c.Writer.Header().Set("Content-Type", "text/html")
		c.String(http.StatusNotFound, string(html))
	})

	Router.Use(static.Serve("/", static.LocalFile(Public, false)))
	Router.Use(middleware.RateLimiter())

	Router.GET("/get", getRoute.GetHandler)
	Router.POST("/parse", parseRoute.ParseHandler)
	Router.GET("/subfinder", subfinderRoute.SubfinderHandler)

	if Port == "" {
		Port = "8080"
	}
	fmt.Println("Server listening on port " + Port)
	Router.Run(":" + Port)
}
