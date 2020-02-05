package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

var (
	ratingApiUrl = "http://192.168.1.155:8000/"
)

func init() {
	if svc, ok := os.LookupEnv("APP_RATING_SVC"); ok {
		ratingApiUrl = svc
	}
}

func addRatingServiceEndpoints(r *gin.Engine) {
	r.POST("/rating/:restaurantId", postRating)
}

func postRating(c *gin.Context) {
	ctx := c.Request.Context()
	restaurantId := c.Param("restaurantId")
	bytes, err := ioutil.ReadAll(c.Request.Body)
	if err != nil {
		c.AbortWithError(http.StatusBadRequest, err)
		return
	}
	ratingStr := string(bytes)
	rating, err := strconv.Atoi(ratingStr)
	if err != nil {
		c.AbortWithError(http.StatusBadRequest, err)
		return
	}
	err = AddRatingToRestaurant(ctx, restaurantId, rating)
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	newRating, err := GetRatingByRestaurantId(ctx, restaurantId)
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	c.Writer.WriteString(fmt.Sprintf("%v", *newRating))
}

func GetRatingByRestaurantId(ctx context.Context, restaurantId string) (*float64, error) {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	url, err := getUrl(ratingApiUrl, "ratings", restaurantId)
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

	var ratings struct {
		Rating *float64 `json:"rating"`
	}
	json.NewDecoder(resp.Body).Decode(&ratings)
	return ratings.Rating, nil
}

func AddRatingToRestaurant(ctx context.Context, restaurantId string, rating int) error {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	url, err := getUrl(ratingApiUrl, "ratings", restaurantId)
	if err != nil {
		return err
	}
	req, err := http.NewRequestWithContext(ctx, "POST", url, strings.NewReader(fmt.Sprint(rating)))
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
