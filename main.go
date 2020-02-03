package main

import (
	"context"
	"github.com/gin-contrib/cors"
	"github.com/gin-contrib/gzip"
	"github.com/gin-gonic/gin"
	"go.undefinedlabs.com/scopeagent/agent"
	"go.undefinedlabs.com/scopeagent/instrumentation/nethttp"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"path"
	"syscall"
	"time"
)

func main() {
	log.SetFlags(log.LstdFlags | log.Lmicroseconds | log.Llongfile)
	nethttp.PatchHttpDefaultClient()
	scopeAgent, err := agent.NewAgent(agent.WithSetGlobalTracer(), agent.WithDebugEnabled())
	if err != nil {
		panic(err)
	}
	defer scopeAgent.Stop()

	log.Println("Starting server...")
	gin.SetMode(gin.ReleaseMode)
	r := gin.Default()
	r.Use(cors.Default())
	r.Use(gzip.Gzip(gzip.DefaultCompression))
	r.GET("/", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})
	addImageServiceEndpoints(r)

	srv := &http.Server{
		Addr:    ":8081",
		Handler: nethttp.Middleware(r),
	}

	go func() {
		log.Println("Listening...")
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %s\n", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutdown Server ...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("Server Shutdown: ", err)
	}

	log.Println("Server exiting")
}

func getUrl(base string, pathValues ...string) (string, error) {
	url, err := url.Parse(base)
	if err != nil {
		return "", err
	}
	args := append([]string{}, url.Path)
	args = append(args, pathValues...)
	url.Path = path.Join(args...)
	return url.String(), nil
}
