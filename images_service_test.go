package main

import (
	"bytes"
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
	test.Run("all-by-restaurant-demo", func(t *testing.T) {
		ctx := scopeagent.GetContextFromTest(t)
		t.Log("getting all images by restaurant")

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
		if ctx.Err() != nil {
			t.Fatal(ctx.Err())
		}
		t.Log("all ok")
	})

	for _, img := range images {
		test.Run("get-demo", func(t *testing.T) {
			ctx := scopeagent.GetContextFromTest(t)
			t.Log("getting an image of a restaurant")

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
			if ctx.Err() != nil {
				t.Fatal(ctx.Err())
			}
			t.Log("all ok")
		})
	}

	var imageId string
	test.Run("post-demo", func(t *testing.T) {
		ctx := scopeagent.GetContextFromTest(t)
		t.Log("posting an image of a restaurant")

		url := fmt.Sprintf("/restaurants/%s/images", restaurantId)
		req, _ := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader([]byte{0, 1, 2, 3}))
		req.Header.Add("Content-Type", "image/custom")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		res := w.Result()

		if res.StatusCode != http.StatusOK {
			t.Fatalf("server: %s respond: %d: %s", url, res.StatusCode, res.Status)
		}
		json.NewDecoder(res.Body).Decode(&imageId)
		if imageId == "" {
			t.Fatal("imageId is nil")
		}
		if ctx.Err() != nil {
			t.Fatal(ctx.Err())
		}
		t.Log("all ok")
	})

	if imageId != "" {
		test.Run("delete-demo", func(t *testing.T) {
			ctx := scopeagent.GetContextFromTest(t)
			t.Log("deleting an image of a restaurant")

			url := fmt.Sprintf("/images/%s", imageId)
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

}
