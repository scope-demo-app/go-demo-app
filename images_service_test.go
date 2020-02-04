package main

import (
	"encoding/json"
	"fmt"
	"go.undefinedlabs.com/scopeagent"
	"net/http"
	"net/http/httptest"
	"testing"
)

const restaurantId = "03d207b0-8015-4ab8-950b-8155b87e1654"

func TestImagesService(t *testing.T) {
	test := scopeagent.GetTest(t)

	var images []string
	test.Run("GetImagesByRestaurant", func(t *testing.T) {
		ctx := scopeagent.GetContextFromTest(t)

		router := setupRouter()
		url := fmt.Sprintf("/restaurants/%s/images", restaurantId)
		req, _ := http.NewRequestWithContext(ctx, "GET", url, nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		res := w.Result()

		if res.StatusCode != http.StatusOK {
			t.Fatalf("server: %s respond: %d: %s", url, res.StatusCode, res.Status)
		}
		json.NewDecoder(res.Body).Decode(&images)
		if images == nil {
			t.Fatal("images can't be nil")
		}
	})

	for _, img := range images {
		t.Run("GetImage:"+img, func(t *testing.T) {
			ctx := scopeagent.GetContextFromTest(t)

			router := setupRouter()
			url := fmt.Sprintf("/images/%s", img)
			req, _ := http.NewRequestWithContext(ctx, "GET", url, nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)
			res := w.Result()

			if res.StatusCode != http.StatusOK {
				t.Fatalf("server: %s respond: %d: %s", url, res.StatusCode, res.Status)
			}

			if res.ContentLength == 0 {
				t.Fatal("content length is nil")
			}
		})
	}
}
