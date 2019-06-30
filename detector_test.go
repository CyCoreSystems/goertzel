package goertzel

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestDetectTone(t *testing.T) {

	found, err := testDetectToneFromFile("1400hz3s.slin", 1400.0, 8000.0, time.Second)
	assert.True(t, found, "1400Hz tone should be found")
	assert.Nil(t, err, "no error should be returned")

	found, err = testDetectToneFromFile("1400hz3s.slin", 1400.0, 8000.0, 5*time.Second)
	assert.False(t, found, "1400Hz tone should be NOT found")
	assert.Nil(t, err, "no error should be returned")

	found, err = testDetectToneFromFile("1400hz3s.slin", 2300.0, 8000.0, 50*time.Millisecond)
	assert.False(t, found, "2300Hz tone should NOT be found")
	assert.Nil(t, err, "no error should be returned")

	found, err = testDetectToneFromFile("1400hz3s.slin", -2300.0, 8000.0, 50*time.Millisecond)
	assert.True(t, found, "2300Hz tone absence should be found")
	assert.Nil(t, err, "no error should be returned")

	found, err = testDetectToneFromFile("combo15s.slin", 2300.0, 8000.0, time.Second)
	assert.True(t, found, "2300Hz tone should be found")
	assert.Nil(t, err, "no error should be returned")

	found, err = testDetectToneFromFile("combo15s.slin", 500.0, 8000.0, time.Second)
	assert.False(t, found, "500Hz tone should NOT be found")
	assert.Nil(t, err, "no should be returned")
}

func testDetectToneFromFile(fn string, freq, rate float64, minDur time.Duration) (bool, error) {
	f, err := os.Open("test/" + fn)
	if err != nil {
		return false, err
	}
	defer f.Close() // nolint

	return DetectTone(context.Background(), freq, rate, minDur, f)
}

func TestDetectToneCancel(t *testing.T) {
	f, err := os.Open("test/combo15s.slin")
	if err != nil {
		t.Errorf("failed to open test file %s: %v", "combo15s.slin", err)
		t.SkipNow()
		return
	}
	defer f.Close() // nolint

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	found, err := DetectTone(ctx, 500.0, 8000.0, time.Second, f)
	assert.False(t, found, "500Hz tone should NOT be found")
	assert.Nil(t, err, "no error should be returned")
}
