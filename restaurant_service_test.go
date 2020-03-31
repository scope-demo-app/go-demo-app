package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"go.undefinedlabs.com/scopeagent"
	"math/rand"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestRestaurantService(t *testing.T) {
	test := scopeagent.GetTest(t)

	test.Run("demotest-all", func(t *testing.T) {
		ctx := scopeagent.GetContextFromTest(t)

		t.Log("getting all restaurants")
		url := "/restaurants"
		req, _ := http.NewRequestWithContext(ctx, "GET", url, nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		res := w.Result()

		if res.StatusCode != http.StatusOK {
			t.Fatalf("server: %s respond: %d: %s", url, res.StatusCode, res.Status)
		}

		var resPayload []restaurant
		err := json.NewDecoder(res.Body).Decode(&resPayload)
		if err != nil {
			t.Fatal(err)
		}
		if ctx.Err() != nil {
			t.Fatal(ctx.Err())
		}
		t.Log("all ok")
	})

	var rsPayload restaurantApi
	test.Run("demotest-create", func(t *testing.T) {
		ctx := scopeagent.GetContextFromTest(t)

		rqPayload := restaurantApiPost{
			Name:        "TestName",
			Description: "TestDescription",
		}
		rqPayloadJson, _ := json.Marshal(rqPayload)

		t.Log("creating restaurant")

		url := "/restaurants"
		req, _ := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(rqPayloadJson))
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		res := w.Result()

		if res.StatusCode != http.StatusOK {
			t.Fatalf("server: %s respond: %d: %s", url, res.StatusCode, res.Status)
		}

		err := json.NewDecoder(res.Body).Decode(&rsPayload)
		if err != nil {
			t.Fatal(err)
		}
		if ctx.Err() != nil {
			t.Fatal(ctx.Err())
		}
		t.Log(rsPayload)
	})

	test.Run("demotest-get", func(t *testing.T) {
		ctx := scopeagent.GetContextFromTest(t)
		t.Log("getting restaurant")

		url := fmt.Sprintf("/restaurants/00000000-0000-0000-0000-000000000001", rsPayload.Id)
		req, _ := http.NewRequestWithContext(ctx, "GET", url, nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		res := w.Result()

		if res.StatusCode != http.StatusOK {
			t.Fatalf("server: %s respond: %d: %s", url, res.StatusCode, res.Status)
		}

		var resPayload restaurant
		err := json.NewDecoder(res.Body).Decode(&resPayload)
		if err != nil {
			t.Fatal(err)
		}
		if ctx.Err() != nil {
			t.Fatal(ctx.Err())
		}
		t.Log(resPayload)
	})

	test.Run("delete", func(t *testing.T) {
		ctx := scopeagent.GetContextFromTest(t)
		t.Log("deleting restaurant")

		url := fmt.Sprintf("/restaurants/%s", rsPayload.Id)
		req, _ := http.NewRequestWithContext(ctx, "DELETE", url, nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		res := w.Result()

		if res.StatusCode != http.StatusOK {
			t.Fatalf("server: %s respond: %d: %s", url, res.StatusCode, res.Status)
		}
		if ctx.Err() != nil {
			t.Fatal(ctx.Err())
		}

		t.Log("all ok")
	})
}

func TestDummySlowBasicEmpty(t *testing.T) {
	test := scopeagent.GetTest(t)
	idx := 0
	for i := 0; i < 200; i++ {
		idx++
		test.Run(fmt.Sprintf("%d", idx), func(t *testing.T) {
			t.Parallel()
			time.Sleep(time.Duration(rand.Intn(10000)) * time.Millisecond)
		})
	}
}

func TestDummyQuickBasicEmpty(t *testing.T) {
	test := scopeagent.GetTest(t)
	idx := 0
	for j := 0; j < 50; j++ {
		idx++
		test.Run(fmt.Sprintf("%d", idx), func(t *testing.T) {
			t.Parallel()
			time.Sleep(time.Duration(rand.Intn(1000)) * time.Millisecond)
		})
	}
}

var benchdata = restaurantApi{
	restaurantApiPost: restaurantApiPost{
		Name:        "TestName",
		Description: "TestDescription",
	},
	Id:        "1234567890",
	Latitude:  nil,
	Longitude: nil,
}

func BenchmarkJsonEncoding(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, _ = json.Marshal(&benchdata)
	}
}
func BenchmarkJsonEncodingWithIndent(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, _ = json.MarshalIndent(&benchdata, "", "  ")
	}
}
