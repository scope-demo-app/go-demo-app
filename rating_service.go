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
)

var (
	ratingApiUrl = "http://localhost:8080"
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
	}
	ratingStr := string(bytes)
	rating, err := strconv.Atoi(ratingStr)
	if err != nil {
		c.AbortWithError(http.StatusBadRequest, err)
	}
	err = AddRatingToRestaurant(ctx, restaurantId, rating)
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
	}
}

func GetRatingByRestaurantId(ctx context.Context, restaurantId string) (*float64, error) {
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
		return nil, errors.New("server didn't respond OK")
	}

	var ratings struct {
		Rating *float64 `json:"rating"`
	}
	json.NewDecoder(resp.Body).Decode(&ratings)
	return ratings.Rating, nil
}

func AddRatingToRestaurant(ctx context.Context, restaurantId string, rating int) error {
	url, err := getUrl(ratingApiUrl, "ratings", restaurantId)
	if err != nil {
		return err
	}
	req, err := http.NewRequestWithContext(ctx, "post", url, strings.NewReader(fmt.Sprint(rating)))
	if err != nil {
		return err
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusAccepted {
		return errors.New("server didn't respond OK")
	}
	return nil
}
