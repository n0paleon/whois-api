package utils

import (
	"math/rand"
	"time"
)

func RandomInRange(min, max int) int {
	source := rand.NewSource(time.Now().UnixNano())
	r := rand.New(source)

	return r.Intn(max-min) + min
}
