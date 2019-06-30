package goertzel

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestOptimalBlockSize(t *testing.T) {
	f := 1400.0
	r := RateTelephony
	min := (50 * time.Second) / 1000
	n := optimalBlockSize(f, r, min)
	assert.Equal(t, 400, n, "optimum block size should match reference")

	f = 1400.0
	r = RateTelephony
	min = (50 * time.Second) / 1000
	n = optimalBlockSize2(f, r, min)
	assert.Equal(t, 160, n, "optimum block size should match reference")

	f = 2300.0
	r = RateTelephony
	min = (50 * time.Second) / 1000
	n = optimalBlockSize(f, r, min)
	assert.Equal(t, 400, n, "optimum block size should match reference")
}

/*
func TestOptimalBlockSizeMultiFreq(t *testing.T) {

	o := &Options{
		Rate:        RateTelephony,
		Frequencies: []float64{1400, 2300},
		MinDuration: 50,
	}
	n := optimalBlockSizeMultiFreq(o)

	// 400 is the best N; 320 is next; base common multiple is 80
	assert.Equal(t, n, 400, "optimum block size should match reference")

	o = &Options{
		Rate:        RateTelephony,
		Frequencies: []float64{697, 770, 852, 941, 1209, 1336, 1477, 1633},
		MinDuration: 250,
	}
	n = optimalBlockSizeMultiFreq(o)

	// No common multiple; use maxN
	assert.Equal(t, n, 2000, "optimum block size should match reference")
}
*/
