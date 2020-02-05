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
	router := setupRouter()

	test.Run("Get", func(t *testing.T) {
		ctx := scopeagent.GetContextFromTest(t)

		rating, err := GetRatingByRestaurantId(ctx, restaurantId)
		if err != nil {
			t.Fatal(err)
		}

		if rating != nil {
			fmt.Println(*rating)
		}
	})

	test.Run("Add", func(t *testing.T) {
		ctx := scopeagent.GetContextFromTest(t)

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
	})
}
