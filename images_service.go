package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"io/ioutil"
	"net/http"
	"os"
	"time"
)

var (
	imagesApiUrl = "https://csharp-demo-app.undefinedlabs.dev/"
)

func init() {
	if svc, ok := os.LookupEnv("APP_IMAGES_SVC"); ok {
		imagesApiUrl = svc
	}
}

func addImageServiceEndpoints(r *gin.Engine) {
	r.GET("/images/:imageId", getImage)
	r.DELETE("/images/:imageId", deleteImage)
	r.GET("/restaurants/:restaurantId/images", getRestaurantImages)
	r.POST("/restaurants/:restaurantId/images", postRestaurantImage)
}

func getImage(c *gin.Context) {
	ctx := c.Request.Context()
	imageId := c.Param("imageId")
	cType, body, err := GetImage(ctx, imageId)
	if err != nil {
		c.AbortWithStatus(http.StatusNotFound)
		panic(err)
	}
	c.Data(http.StatusOK, cType, body)
}

func deleteImage(c *gin.Context) {
	ctx := c.Request.Context()
	imageId := c.Param("imageId")
	err := DeleteImage(ctx, imageId)
	if err != nil {
		c.AbortWithError(http.StatusBadRequest, err)
		panic(err)
	}
	c.Status(http.StatusOK)
}

func getRestaurantImages(c *gin.Context) {
	ctx := c.Request.Context()
	restaurantId := c.Param("restaurantId")
	values, err := GetImagesByRestaurant(ctx, restaurantId)
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		panic(err)
	}
	c.JSON(http.StatusOK, values)
}

func postRestaurantImage(c *gin.Context) {
	ctx := c.Request.Context()
	restaurantId := c.Param("restaurantId")

	bytes, _ := ioutil.ReadAll(c.Request.Body)
	value, err := AddImageToRestaurant(ctx, restaurantId, c.Request.Header.Get("Content-Type"), bytes)
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		panic(err)
	}

	c.JSON(http.StatusOK, value)
}

//

func GetImagesByRestaurant(ctx context.Context, restaurantId string) ([]string, error) {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	url, err := getUrl(imagesApiUrl, "images", "restaurant", restaurantId)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, errors.New(fmt.Sprintf("server: %s respond: %d: %s", url, resp.StatusCode, resp.Status))
	}
	var images []string
	json.NewDecoder(resp.Body).Decode(&images)
	return images, nil
}

func AddImageToRestaurant(ctx context.Context, restaurantId string, contentType string, data []byte) (string, error) {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	url, err := getUrl(imagesApiUrl, "images", "restaurant", restaurantId)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(data))
	if err != nil {
		return "", err
	}
	req.Header.Add("Content-Type", contentType)
	req.Header.Add("Content-Length", fmt.Sprint(len(data)))

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}

	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusAccepted {
		return "", errors.New(fmt.Sprintf("server: %s respond: %d: %s", url, resp.StatusCode, resp.Status))
	}
	var imageId string
	json.NewDecoder(resp.Body).Decode(&imageId)
	if imageId == "" {
		return "", errors.New("image could not be uploaded")
	}
	return imageId, nil
}

func DeleteImagesByRestaurant(ctx context.Context, restaurantId string) error {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	imgs, err := GetImagesByRestaurant(ctx, restaurantId)
	if err != nil {
		return err
	}
	var lastError error
	for idx := range imgs {
		lastError = DeleteImage(ctx, imgs[idx])
	}
	return lastError
}

func GetImage(ctx context.Context, imageId string) (string, []byte, error) {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	url, err := getUrl(imagesApiUrl, "images", imageId)
	if err != nil {
		return "", nil, err
	}

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return "", nil, err
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", nil, errors.New("server didn't respond OK")
	}

	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", nil, err
	}
	contentType := resp.Header.Get("Content-type")
	if contentType == "" {
		contentType = http.DetectContentType(bodyBytes)
	}
	return contentType, bodyBytes, nil
}

func DeleteImage(ctx context.Context, imageId string) error {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	url, err := getUrl(imagesApiUrl, "images", imageId)
	if err != nil {
		return err
	}
	req, err := http.NewRequestWithContext(ctx, "DELETE", url, nil)
	if err != nil {
		return err
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusAccepted {
		return errors.New(fmt.Sprintf("server: %s respond: %d: %s", url, resp.StatusCode, resp.Status))
	}
	return nil
}
