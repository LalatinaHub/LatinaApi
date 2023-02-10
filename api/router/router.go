package router

import (
	"fmt"
	"net/http"
	"os"

	getRoute "github.com/LalatinaHub/LatinaApi/api/get"
	logRoute "github.com/LalatinaHub/LatinaApi/api/log"
	"github.com/gin-contrib/static"
	"github.com/gin-gonic/gin"
)

var (
	Router = gin.Default()
	Port   = os.Getenv("PORT")
)

func Start() {
	Router.SetTrustedProxies(nil)

	Router.NoRoute(func(c *gin.Context) {
		html, _ := os.ReadFile("public/404.html")

		c.Writer.Header().Set("Content-Type", "text/html")
		c.String(http.StatusNotFound, string(html))
	})

	Router.GET("/get", getRoute.GetHandler)
	Router.GET("/log", logRoute.LogHandler)

	Router.Use(static.Serve("/", static.LocalFile("public/", false)))

	if Port == "" {
		Port = "8080"
	}
	fmt.Println("Server listening on port " + Port)
	Router.Run(":" + Port)
}
