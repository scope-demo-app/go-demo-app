package main

import (
	"github.com/gin-gonic/gin"
	"go.undefinedlabs.com/scopeagent"
	"go.undefinedlabs.com/scopeagent/agent"
	"go.undefinedlabs.com/scopeagent/instrumentation/nethttp"
	"log"
	"math/rand"
	"os"
	"testing"
	"time"
)

var router *gin.Engine

func TestMain(m *testing.M) {
	log.SetFlags(log.LstdFlags | log.Lmicroseconds | log.Llongfile)
	nethttp.PatchHttpDefaultClient(nethttp.WithPayloadInstrumentation())
	rand.Seed(time.Now().UnixNano())
	router = setupRouter()
	os.Exit(scopeagent.Run(m, agent.WithSetGlobalTracer(), agent.WithDebugEnabled(), agent.WithRetriesOnFail(3)))
}

func setupRouter() *gin.Engine {
	r := gin.Default()
	r.Use(logErrorOnSpanMiddleware)
	r.Use(errorInjectionMiddleware)
	addImageServiceEndpoints(r)
	addRatingServiceEndpoints(r)
	addRestaurantServiceEndpoints(r)
	return r
}
