package main

import (
	"github.com/gin-gonic/gin"
	"go.undefinedlabs.com/scopeagent"
	"go.undefinedlabs.com/scopeagent/agent"
	"go.undefinedlabs.com/scopeagent/instrumentation/nethttp"
	"log"
	"os"
	"testing"
)

func TestMain(m *testing.M) {
	log.SetFlags(log.LstdFlags | log.Lmicroseconds | log.Llongfile)
	nethttp.PatchHttpDefaultClient(nethttp.WithPayloadInstrumentation())
	os.Exit(scopeagent.Run(m, agent.WithSetGlobalTracer(), agent.WithDebugEnabled()))
}

func setupRouter() *gin.Engine {
	r := gin.Default()
	addImageServiceEndpoints(r)
	addRatingServiceEndpoints(r)
	addRestaurantServiceEndpoints(r)
	return r
}
