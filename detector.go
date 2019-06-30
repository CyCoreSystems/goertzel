package goertzel

import (
	"context"
	"io"
	"math"
	"time"

	"github.com/pkg/errors"
)

func detectTone(pCtx context.Context, freq, sampleRate float64, minDuration time.Duration, in io.Reader) (sufficient bool, err error) {
	var findingAbsence bool
	if freq < 0 {
		findingAbsence = true
		freq = -freq
	}

	t := NewTarget(freq, sampleRate, minDuration)
	defer t.Stop()

	ctx, cancel := context.WithCancel(pCtx)
	defer cancel()

	go func() {
		if err = t.Read(in); err != nil {
			if err == io.EOF {
				return
			}
			err = errors.Wrap(err, "error reading from input")
		}
	}()

	// Figure out the time-size of each block to determine the blocks required to constitute detection
	timeSize := float64(t.blockSize) / sampleRate
	reqBlocks := int(math.Ceil(minDuration.Seconds() / float64(timeSize)))

	var count int
	var found bool
	for b := range t.Blocks() {
		if findingAbsence {
			found = !b.Present
		} else {
			found = b.Present
		}

		if found {
			count++
		} else {
			count = 0
		}

		if count >= reqBlocks {
			return true, nil
		}

		if ctx.Err() != nil {
			return
		}
	}
	return

}

// DetectTone waits for the given tone to be found, returning with `true` when it is.  `false` will be returned if canceled by context or by a stream error/completion.
func DetectTone(ctx context.Context, freq, sampleRate float64, minDuration time.Duration, in io.Reader) (found bool, err error) {
	return detectTone(ctx, freq, sampleRate, minDuration, in)
}

// DetectToneAbsence waits for the given frequency to go away for the requested amount of time
func DetectToneAbsence(ctx context.Context, freq, sampleRate float64, minDuration time.Duration, in io.Reader) (found bool, err error) {
	return detectTone(ctx, -freq, sampleRate, minDuration, in)
}
