package api

import (
	"net/http"
	"net/http/pprof"

	"github.com/LalatinaHub/LatinaApi/internal/api/handler"
	"github.com/LalatinaHub/LatinaApi/internal/api/middleware"
	"github.com/gin-gonic/gin"
)

// RouterConfig contains all dependencies required to initialize the Gin router.
type RouterConfig struct {
	SubHandler     *handler.SubHandler
	UserHandler    *handler.UserHandler
	AdminHandler   *handler.AdminHandler
	HealthHandler  *handler.HealthHandler
	InfoHandler    *handler.InfoHandler
	ConvertHandler *handler.ConvertHandler
	RateLimiter    *middleware.RateLimiter
	IsProduction   bool
}

// SetupRouter creates and configures the Gin engine with all routes and middlewares.
func SetupRouter(cfg RouterConfig) *gin.Engine {
	if cfg.IsProduction {
		gin.SetMode(gin.ReleaseMode)
	} else {
		gin.SetMode(gin.DebugMode)
	}

	r := gin.New()

	// Global middlewares
	r.Use(gin.Recovery())
	r.Use(middleware.RequestLoggerMiddleware())
	r.Use(middleware.CORSMiddleware())
	r.Use(middleware.ErrorMiddleware())

	// Health and ping (not rate limited)
	r.GET("/health", cfg.HealthHandler.Health)
	r.GET("/ping", cfg.HealthHandler.Ping)

	// Profiling endpoints (disabled in production)
	if !cfg.IsProduction {
		pprofGroup := r.Group("/debug/pprof")
		{
			pprofGroup.GET("/", gin.WrapF(pprof.Index))
			pprofGroup.GET("/cmdline", gin.WrapF(pprof.Cmdline))
			pprofGroup.GET("/profile", gin.WrapF(pprof.Profile))
			pprofGroup.GET("/symbol", gin.WrapF(pprof.Symbol))
			pprofGroup.POST("/symbol", gin.WrapF(pprof.Symbol))
			pprofGroup.GET("/trace", gin.WrapF(pprof.Trace))
			pprofGroup.GET("/allocs", gin.WrapH(pprof.Handler("allocs")))
			pprofGroup.GET("/block", gin.WrapH(pprof.Handler("block")))
			pprofGroup.GET("/goroutine", gin.WrapH(pprof.Handler("goroutine")))
			pprofGroup.GET("/heap", gin.WrapH(pprof.Handler("heap")))
			pprofGroup.GET("/mutex", gin.WrapH(pprof.Handler("mutex")))
			pprofGroup.GET("/threadcreate", gin.WrapH(pprof.Handler("threadcreate")))
		}
	}

	// Root info
	r.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"service": "LatinaApi",
			"status":  "running",
		})
	})

	// Subscription route at root
	r.GET("/sub", cfg.RateLimiter.Middleware(), cfg.SubHandler.HandleSub)

	// API v1 grouping
	apiV1 := r.Group("/api/v1")
	apiV1.Use(cfg.RateLimiter.Middleware())
	{
		apiV1.GET("/ping", cfg.HealthHandler.Ping)
		apiV1.GET("/info", cfg.InfoHandler.Info)
		apiV1.GET("/sub", cfg.SubHandler.HandleSub)
		apiV1.POST("/convert", cfg.ConvertHandler.Convert)
	}

	// User management group
	userGroup := r.Group("/user")
	userGroup.Use(cfg.RateLimiter.Middleware())
	{
		userGroup.GET("/:apiToken/:id", cfg.UserHandler.GetUser)
	}

	// DB admin group
	dbGroup := r.Group("/db")
	dbGroup.Use(cfg.RateLimiter.Middleware())
	{
		dbGroup.POST("/:apiToken/exec", cfg.AdminHandler.ExecSQL)
	}

	return r
}
