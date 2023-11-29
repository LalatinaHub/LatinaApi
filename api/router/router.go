package router

import (
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"

	getRoute "github.com/LalatinaHub/LatinaApi/api/get"
	logRoute "github.com/LalatinaHub/LatinaApi/api/log"
	"github.com/LalatinaHub/LatinaApi/api/middleware"
	ocrRoute "github.com/LalatinaHub/LatinaApi/api/ocr"
	parseRoute "github.com/LalatinaHub/LatinaApi/api/parse"
	subfinderRoute "github.com/LalatinaHub/LatinaApi/api/subfinder"
	"github.com/gin-gonic/gin"
)

var (
	Router = gin.Default()
	Port   = os.Getenv("PORT")
	Public = "web/public/"
)

func Start() {
	Router.SetTrustedProxies(nil)

	// Penggunaan middleware RateLimiter
	Router.Use(middleware.RateLimiter())

	// Route yang digunakan pada aplikasi
	Router.GET("/get", getRoute.GetHandler)
	Router.GET("/log", logRoute.LogHandler)
	Router.POST("/parse", parseRoute.ParseHandler)
	Router.GET("/subfinder", subfinderRoute.SubfinderHandler)
	Router.GET("/ocr", ocrRoute.OcrHandler)

	// Middleware untuk reverse proxy
	Router.NoRoute(ReverseProxy)

	// Menetapkan port untuk server
	if Port == "" {
		Port = "8080"
	}
	fmt.Println("Server listening on port " + Port)
	Router.Run(":" + Port)
}

// Middleware ReverseProxy untuk meneruskan permintaan ke backend
func ReverseProxy(c *gin.Context) {
	remote, err := url.Parse("http://lalatinahub.github.io/")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal Server Error"})
		return
	}

	proxy := httputil.NewSingleHostReverseProxy(remote)
	proxy.Director = func(req *http.Request) {
		req.Header = c.Request.Header // Opsional, tergantung kebutuhan
		req.Host = remote.Host
		req.URL.Scheme = remote.Scheme
		req.URL.Host = remote.Host
		req.URL.Path = fmt.Sprintf("/LatinaDocs%s", c.Request.URL.Path)
	}

	proxy.ServeHTTP(c.Writer, c.Request)
}
