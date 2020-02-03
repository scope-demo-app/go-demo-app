package main

import "github.com/gin-gonic/gin"

var (
	ratingApiUrl = "http://localhost:8080"
)

func addRatingServiceEndpoints(r *gin.Engine) {
	r.GET("/restaurant/:restaurantId/ratings", getRestaurantRatings)
	r.POST("/restaurant/:restaurantId/ratings", postRestaurantRating)
}

func getRestaurantRatings(c *gin.Context) {
}

func postRestaurantRating(c *gin.Context) {
}
