package goertzel

import (
	"io"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestTargetBlockSummaryReading(t *testing.T) {
	mag, found := testBlockSummaryReading(t, "1400hz3s.slin", 1400.0)
	assert.True(t, found, "1400Hz tone should be found")

	mag, found = testBlockSummaryReading(t, "1400hz3s.slin", 2300.0)
	if mag > ToneThreshold {
		t.Errorf("magnitude of 2300Hz tone detection (%f) too high", mag)
	}

	mag, found = testBlockSummaryReading(t, "2300hz3s.slin", 2300.0)
	if mag < ToneThreshold {
		t.Errorf("magnitude of 2300Hz tone detection (%f) too low", mag)
	}

	mag, found = testBlockSummaryReading(t, "silence3s.slin", 500.0)
	if mag > ToneThreshold {
		t.Errorf("magnitude of 500Hz tone detection (%f) too high", mag)
	}
	assert.False(t, found, "500Hz tone should NOT be found")

	_, found = testBlockSummaryReading(t, "combo15s.slin", 1400.0)
	assert.True(t, found, "1400Hz tone should be found")

	_, found = testBlockSummaryReading(t, "combo15s.slin", 2300.0)
	assert.True(t, found, "2300Hz tone should be found")

	_, found = testBlockSummaryReading(t, "combo15s.slin", 500.0)
	assert.False(t, found, "500Hz tone should NOT be found")

	// Close enough test
	_, found = testBlockSummaryReading(t, "2310hz3s.slin", 2300.0)
	assert.True(t, found, "2300Hz tone should be found from 2310Hz")

	// Just outside test
	_, found = testBlockSummaryReading(t, "2350hz3s.slin", 2300.0)
	assert.False(t, found, "2300Hz tone should NOT be found from 2350Hz")
}

func testBlockSummaryReading(t *testing.T, fn string, freq float64) (float64, bool) {
	var highestMag float64
	var found bool

	tgt := NewTarget(freq, RateTelephony, 50*time.Millisecond)
	tgt.UseOptimized = false
	defer tgt.Stop()

	go func() {
		var i int
		for b := range tgt.Blocks() {
			i++
			//log.Printf("received block %d: %+v", i, b)
			assert.Equal(t, b.Samples, tgt.blockSize, "block size should match reference")
			if b.Magnitude2 > highestMag {
				highestMag = b.Magnitude2
			}
			if b.Present {
				found = true
			}
		}
		//t.Logf("processed %d blocks from %s for %fHz detection", i, fn, freq)
	}()

	f, err := os.Open("test/" + fn)
	if err != nil {
		t.Fatalf("failed to open test file %s: %v", fn, err)
	}
	defer f.Close()
	if err := tgt.ingest(f); err != nil {
		if err != io.EOF {
			t.Errorf("failed to ingest test file %s: %v", fn, err)
		}
	}
	tgt.Stop()

	return highestMag, found
}
