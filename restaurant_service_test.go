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
)

func TestRestaurantService(t *testing.T) {
	test := scopeagent.GetTest(t)
	router := setupRouter()
	for i := 0; i < 30; i++ {
		rand.Int()
	}

	test.Run("All", func(t *testing.T) {
		ctx := scopeagent.GetContextFromTest(t)

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
		fmt.Println(resPayload)
	})

	var rsPayload restaurantApi
	test.Run("Create", func(t *testing.T) {
		ctx := scopeagent.GetContextFromTest(t)

		rqPayload := restaurantApiPost{
			Name:        "TestName",
			Description: "TestDescription",
		}
		rqPayloadJson, _ := json.Marshal(rqPayload)

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
		fmt.Println(rsPayload)
	})

	test.Run("Get", func(t *testing.T) {
		ctx := scopeagent.GetContextFromTest(t)

		url := fmt.Sprintf("/restaurants/%s", rsPayload.Id)
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
		fmt.Println(resPayload)
	})

	test.Run("Delete", func(t *testing.T) {
		ctx := scopeagent.GetContextFromTest(t)

		url := fmt.Sprintf("/restaurants/%s", rsPayload.Id)
		req, _ := http.NewRequestWithContext(ctx, "DELETE", url, nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		res := w.Result()

		if res.StatusCode != http.StatusOK {
			t.Fatalf("server: %s respond: %d: %s", url, res.StatusCode, res.Status)
		}
	})
}
