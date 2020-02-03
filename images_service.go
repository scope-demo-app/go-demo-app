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
)

var (
	imagesApiUrl = "http://localhost:8080"
)

func addImageServiceEndpoints(r *gin.Engine) {
	r.GET("/images/:imageId", getImage)
	r.DELETE("/images/:imageId", deleteImage)

	// TODO: delete the following endpoints
	r.GET("/restaurant/:restaurantId/images", getRestaurantImages)
	r.POST("/restaurant/:restaurantId/images", postRestaurantImage)
}

func getImage(c *gin.Context) {
	ctx := c.Request.Context()
	imageId := c.Param("imageId")

	url, err := getUrl(imagesApiUrl, "images", imageId)
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		c.AbortWithError(http.StatusBadRequest, err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		c.AbortWithError(resp.StatusCode, errors.New("server didn't respond OK"))
		return
	}

	bytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	contentType := resp.Header.Get("Content-type")
	if contentType == "" {
		contentType = http.DetectContentType(bytes)
	}
	c.Data(http.StatusOK, contentType, bytes)
}

func deleteImage(c *gin.Context) {
	ctx := c.Request.Context()
	imageId := c.Param("imageId")

	url, err := getUrl(imagesApiUrl, "images", imageId)
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	req, err := http.NewRequestWithContext(ctx, "DELETE", url, nil)
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		c.AbortWithError(http.StatusBadRequest, err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		c.AbortWithError(resp.StatusCode, errors.New("server didn't respond OK"))
		return
	}
	c.Status(http.StatusOK)
}

func getRestaurantImages(c *gin.Context) {
	ctx := c.Request.Context()
	restaurantId := c.Param("restaurantId")
	values, err := getImagesByRestaurant(ctx, restaurantId)
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
	}
	c.JSON(http.StatusOK, values)
}

func postRestaurantImage(c *gin.Context) {
	ctx := c.Request.Context()
	restaurantId := c.Param("restaurantId")

	bytes, _ := ioutil.ReadAll(c.Request.Body)
	value, err := postImageToRestaurant(ctx, restaurantId, c.Request.Header.Get("Content-Type"), bytes)
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
	}

	c.JSON(http.StatusOK, value)
}

func getImagesByRestaurant(ctx context.Context, restaurantId string) ([]string, error) {
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
		return nil, errors.New("server didn't respond OK")
	}
	var images []string
	json.NewDecoder(resp.Body).Decode(&images)
	return images, nil
}

func postImageToRestaurant(ctx context.Context, restaurantId string, contentType string, data []byte) (string, error) {
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
	if resp.StatusCode != http.StatusOK {
		return "", errors.New("server didn't respond OK")
	}
	var imageId string
	json.NewDecoder(resp.Body).Decode(&imageId)
	if imageId == "" {
		return "", errors.New("image could not be uploaded")
	}
	return imageId, nil
}
