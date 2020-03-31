package main

import (
	"fmt"
	"go.undefinedlabs.com/scopeagent"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestRatingService(t *testing.T) {
	test := scopeagent.GetTest(t)

	test.Run("demotest-get", func(t *testing.T) {
		ctx := scopeagent.GetContextFromTest(t)
		t.Log("getting rating")

		rating, err := GetRatingByRestaurantId(ctx, restaurantId)
		if err != nil {
			t.Fatal(err)
		}

		if rating != nil {
			t.Log(*rating)
		}
		if ctx.Err() != nil {
			t.Fatal(ctx.Err())
		}
		t.Log("all ok")
	})

	test.Run("add", func(t *testing.T) {
		ctx := scopeagent.GetContextFromTest(t)
		t.Log("add rating")

		url := fmt.Sprintf("/rating/%s", restaurantId)
		req, _ := http.NewRequestWithContext(ctx, "POST", url, strings.NewReader("4"))
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		res := w.Result()

		if res.StatusCode != http.StatusOK {
			t.Fatalf("server: %s respond: %d: %s", url, res.StatusCode, res.Status)
		}
		bytes, err := ioutil.ReadAll(res.Body)
		if err != nil {
			t.Fatal(err)
		}
		fmt.Println(string(bytes))
		if ctx.Err() != nil {
			t.Fatal(ctx.Err())
		}
		t.Log("all ok")
	})
}
