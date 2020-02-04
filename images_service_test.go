package main

import (
	"go.undefinedlabs.com/scopeagent"
	"testing"
)

const restaurantId = "03d207b0-8015-4ab8-950b-8155b87e1654"

func TestGetImagesByRestaurant(t *testing.T) {
	ctx := scopeagent.GetContextFromTest(t)

	imgs, err := GetImagesByRestaurant(ctx, restaurantId)
	if err != nil {
		panic(err)
	}

	if imgs == nil {
		t.Fatal("images is nil")
	}
}
