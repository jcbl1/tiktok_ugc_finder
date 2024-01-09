package utils

import (
	"fmt"
	"math/rand"
	"time"
)

func ShortInterval() time.Duration {
	itv := rand.Intn(600) + 231
	dur, _ := time.ParseDuration(fmt.Sprintf("%dms", itv))
	return dur
}

func MediumInterval() time.Duration {
	itv := rand.Intn(1997) + 2024
	dur, _ := time.ParseDuration(fmt.Sprintf("%dms", itv))
	return dur
}

func LongInterval() time.Duration {
	itv := rand.Intn(8888) + 7777
	dur, _ := time.ParseDuration(fmt.Sprintf("%dms", itv))
	return dur
}
