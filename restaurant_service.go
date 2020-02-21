package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"math/rand"
	"net/http"
	"os"
	"sync"
	"sync/atomic"
	"time"
)

type (
	restaurant struct {
		restaurantApi
		Rating *float64 `json:"rating"`
		Images []string `json:"images"`
	}

	restaurantApi struct {
		restaurantApiPost
		Id        string  `json:"id,omitempty"`
		Latitude  *string `json:"latitude"`
		Longitude *string `json:"longitude"`
	}

	restaurantApiPost struct {
		Name        string `json:"name"`
		Description string `json:"description"`
	}

	restaurantPost struct {
		restaurantApiPost
		Images *[]restaurantPostImage `json:"images"`
	}

	restaurantPostImage struct {
		MimeType string `json:"mimeType"`
		Data     []byte `json:"data"`
	}
)

var (
	restaurantApiUrl = "https://java-demo-app.undefinedlabs.dev/"
	counter          int64
)

func init() {
	if svc, ok := os.LookupEnv("APP_RESTAURANT_SVC"); ok {
		restaurantApiUrl = svc
	}
}

func addRestaurantServiceEndpoints(r *gin.Engine) {
	r.GET("/restaurants", getRestaurants)
	r.GET("/restaurants/:restaurantId", getRestaurantById)
	r.POST("/restaurants", postRestaurant)
	r.PATCH("/restaurants/:restaurantId", patchRestaurant)
	r.DELETE("/restaurants/:restaurantId", deleteRestaurant)
}

func getRestaurants(c *gin.Context) {
	ctx := c.Request.Context()
	var r []restaurantApi
	var err error
	if c.Query("name") != "" {
		r, err = GetAllRestaurantsByName(ctx, c.Query("name"))
	} else {
		r, err = GetAllRestaurants(ctx)
	}
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		panic(err)
	}
	rests := make([]restaurant, 0)
	var wg sync.WaitGroup
	wg.Add(len(r) * 2)

	for idx := range r {
		rests = append(rests, restaurant{restaurantApi: r[idx]})

		go func(index int) {
			defer wg.Done()

			imgs, err := GetImagesByRestaurant(ctx, r[index].Id)
			if err != nil {
				c.Error(err)
				logError(c, err)
			}
			for _, item := range imgs {
				rests[index].Images = append(rests[index].Images, fmt.Sprintf("/images/%s", item))
			}

		}(idx)

		go func(index int) {
			defer wg.Done()

			rating, err := GetRatingByRestaurantId(ctx, r[index].Id)
			if err != nil {
				c.Error(err)
				logError(c, err)
			}
			rests[index].Rating = rating

		}(idx)

	}

	wg.Wait()
	c.JSON(http.StatusOK, rests)
}

func getRestaurantById(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), getTimeoutDuration())
	defer cancel()
	restaurantId := c.Param("restaurantId")

	var r *restaurantApi
	var rErr error
	var imgs []string
	var imgsErr error
	var rating *float64
	var ratingErr error
	var wg sync.WaitGroup

	wg.Add(3)
	go func() {
		defer wg.Done()
		r, rErr = GetRestaurantById(ctx, restaurantId)
	}()
	go func() {
		defer wg.Done()
		imgs, imgsErr = GetImagesByRestaurant(ctx, restaurantId)
	}()
	go func() {
		defer wg.Done()
		rating, ratingErr = GetRatingByRestaurantId(ctx, restaurantId)
	}()

	wg.Wait()

	if rErr != nil {
		c.AbortWithError(http.StatusInternalServerError, rErr)
		panic(rErr)
	}
	if imgsErr != nil {
		c.Error(imgsErr)
		panic(imgsErr)
	}
	if ratingErr != nil {
		c.Error(ratingErr)
		panic(ratingErr)
	}
	var rest = restaurant{restaurantApi: *r}
	for _, item := range imgs {
		rest.Images = append(rest.Images, fmt.Sprintf("/images/%s", item))
	}
	rest.Rating = rating
	c.JSON(http.StatusOK, rest)
}

func postRestaurant(c *gin.Context) {
	ctx := c.Request.Context()
	var restRq restaurantPost
	err := c.BindJSON(&restRq)
	if err != nil {
		c.AbortWithError(http.StatusBadRequest, err)
		panic(err)
	}
	r, err := AddRestaurant(ctx, restRq.restaurantApiPost)
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		panic(err)
	}
	var rest = restaurant{restaurantApi: *r}
	if restRq.Images != nil {
		for _, item := range *restRq.Images {
			imgId, err := AddImageToRestaurant(ctx, rest.Id, item.MimeType, item.Data)
			if err != nil {
				c.Error(err)
				logError(c, err)
			}
			rest.Images = append(rest.Images, fmt.Sprintf("/images/%s", imgId))
		}
	}
	c.JSON(http.StatusOK, rest)
}

func patchRestaurant(c *gin.Context) {
	ctx := c.Request.Context()
	restaurantId := c.Param("restaurantId")

	var restRq restaurantApi
	err := c.BindJSON(&restRq)
	if err != nil {
		c.AbortWithError(http.StatusBadRequest, err)
		panic(err)
	}

	r, err := UpdateRestaurant(ctx, restaurantId, restRq)
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		panic(err)
	}

	rest := restaurant{restaurantApi: *r}
	imgs, err := GetImagesByRestaurant(ctx, r.Id)
	if err != nil {
		c.Error(err)
		logError(c, err)
	}
	for _, item := range imgs {
		rest.Images = append(rest.Images, fmt.Sprintf("/images/%s", item))
	}
	c.JSON(http.StatusOK, rest)
}

func deleteRestaurant(c *gin.Context) {
	ctx := c.Request.Context()
	restaurantId := c.Param("restaurantId")

	err := DeleteRestaurantById(ctx, restaurantId)
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		panic(err)
	}

	err = DeleteImagesByRestaurant(ctx, restaurantId)
	if err != nil {
		c.Error(err)
		logError(c, err)
	}
}

func GetAllRestaurants(ctx context.Context) ([]restaurantApi, error) {
	url, err := getUrl(restaurantApiUrl, "restaurants")
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
	var rest []restaurantApi
	json.NewDecoder(resp.Body).Decode(&rest)
	return rest, nil
}

func GetRestaurantById(ctx context.Context, restaurantId string) (*restaurantApi, error) {
	url, err := getUrl(restaurantApiUrl, "restaurants", restaurantId)
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
	var rest restaurantApi
	json.NewDecoder(resp.Body).Decode(&rest)
	return &rest, nil
}

func GetAllRestaurantsByName(ctx context.Context, name string) ([]restaurantApi, error) {
	url, err := getUrl(restaurantApiUrl, "restaurants")
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}
	q := req.URL.Query()
	q.Add("name", name)
	req.URL.RawQuery = q.Encode()
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, errors.New(fmt.Sprintf("server: %s respond: %d: %s", url, resp.StatusCode, resp.Status))
	}
	var rest []restaurantApi
	json.NewDecoder(resp.Body).Decode(&rest)
	return rest, nil
}

func DeleteRestaurantById(ctx context.Context, restaurantId string) error {
	url, err := getUrl(restaurantApiUrl, "restaurants", restaurantId)
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

func AddRestaurant(ctx context.Context, post restaurantApiPost) (*restaurantApi, error) {
	url, err := getUrl(restaurantApiUrl, "restaurants")
	if err != nil {
		return nil, err
	}
	postBytes, err := json.Marshal(post)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(postBytes))
	if err != nil {
		return nil, err
	}
	req.Header.Add("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusAccepted {
		return nil, errors.New(fmt.Sprintf("server: %s respond: %d: %s", url, resp.StatusCode, resp.Status))
	}
	var rest restaurantApi
	json.NewDecoder(resp.Body).Decode(&rest)
	return &rest, nil
}

func UpdateRestaurant(ctx context.Context, restaurantId string, post restaurantApi) (*restaurantApi, error) {
	url, err := getUrl(restaurantApiUrl, "restaurants", restaurantId)
	if err != nil {
		return nil, err
	}
	postBytes, err := json.Marshal(post)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequestWithContext(ctx, "PATCH", url, bytes.NewReader(postBytes))
	if err != nil {
		return nil, err
	}
	req.Header.Add("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, errors.New(fmt.Sprintf("server: %s respond: %d: %s", url, resp.StatusCode, resp.Status))
	}
	var rest restaurantApi
	json.NewDecoder(resp.Body).Decode(&rest)
	return &rest, nil
}

func getTimeoutDuration() time.Duration {
	duration := time.Duration(rand.Intn(50)) * time.Millisecond
	c := atomic.AddInt64(&counter, 1)
	if c%2 == 0 {
		duration = time.Duration(rand.Intn(500)) * time.Millisecond
	}
	fmt.Println(duration)
	return duration
}
