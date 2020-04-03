package main

import (
	"context"
	"github.com/gin-contrib/cors"
	"github.com/gin-contrib/gzip"
	"github.com/gin-gonic/gin"
	"github.com/opentracing/opentracing-go"
	"go.undefinedlabs.com/scopeagent/agent"
	"go.undefinedlabs.com/scopeagent/errors"
	"go.undefinedlabs.com/scopeagent/instrumentation/nethttp"
	"log"
	"math/rand"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"path"
	"strconv"
	"syscall"
	"time"
)

var GitCommit string
var GitSourceRoot string

func main() {
	log.SetFlags(log.LstdFlags | log.Lmicroseconds | log.Llongfile)
	rand.Seed(time.Now().UnixNano())
	nethttp.PatchHttpDefaultClient(nethttp.WithPayloadInstrumentation())
	opts := []agent.Option{agent.WithSetGlobalTracer(), agent.WithDebugEnabled()}
	if GitCommit != "" {
		opts = append(opts, agent.WithGitInfo("", GitCommit, GitSourceRoot))
	}
	scopeAgent, err := agent.NewAgent(opts...)
	if err != nil {
		panic(err)
	}
	defer scopeAgent.Stop()

	log.Println("Starting server...")
	gin.SetMode(gin.ReleaseMode)
	r := gin.Default()
	r.Use(cors.New(cors.Config{
		AllowAllOrigins:        true,
		AllowMethods:           []string{"GET", "POST", "PUT", "PATCH", "DELETE", "HEAD"},
		AllowHeaders:           []string{
			"Origin",
			"Content-Length",
			"Content-Type",
			"ot-tracer-traceid",
			"ot-tracer-spanid",
			"ot-tracer-parentspanid",
			"ot-tracer-sampled",
			"ot-baggage-trace.kind",
			"traceparent",
			"tracestate",
		},
		AllowCredentials:       true,
		MaxAge:                 30 * time.Second,
		AllowWildcard:          true,
		AllowBrowserExtensions: true,
		AllowWebSockets:        true,
		AllowFiles:             true,
	}))
	r.Use(logErrorOnSpanMiddleware)
	r.Use(errorInjectionMiddleware)
	r.Use(gzip.Gzip(gzip.DefaultCompression))
	r.GET("/", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})
	addImageServiceEndpoints(r)
	addRatingServiceEndpoints(r)
	addRestaurantServiceEndpoints(r)
	srv := &http.Server{
		Addr:    ":80",
		Handler: nethttp.Middleware(r, nethttp.MWPayloadInstrumentation()),
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
func errorInjectionMiddleware(c *gin.Context) {
	const keySleep = "rs.sleep"
	const keyStatus = "rs.status"
	const keyFailurePercentage = "rs.failure"

	qKeySleep := c.Query(keySleep)
	qKeyStatus := c.Query(keyStatus)
	qKeyFailurePercentage := c.Query(keyFailurePercentage)

	if qKeySleep != "" {
		if sleepValue, err := strconv.Atoi(qKeySleep); err == nil {
			<-time.After(time.Duration(sleepValue) * time.Millisecond)
		}
	}

	if qKeyStatus != "" {
		if statusValue, err := strconv.Atoi(qKeyStatus); err == nil {
			c.AbortWithStatus(statusValue)
			return
		}
	}

	if qKeyFailurePercentage != "" {
		if failurePercentage, err := strconv.Atoi(qKeyFailurePercentage); err == nil {
			if rand.Intn(100) <= failurePercentage {
				//c.AbortWithStatus(http.StatusInternalServerError)
				//return
				panic("error processing request.")
			}
		}
	}

	c.Next()
}

func logErrorOnSpanMiddleware(c *gin.Context) {
	defer func() {
		if r := recover(); r != nil {
			errors.LogPanic(c.Request.Context(), r, 1)
			panic(r)
		}
	}()
	c.Next()
}

func logError(c *gin.Context, err error) {
	sp := opentracing.SpanFromContext(c.Request.Context())
	if sp != nil {
		errors.LogPanic(c.Request.Context(), err, 1)
		sp.SetTag("error", false)
	}
}
