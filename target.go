package goertzel

import (
	"bufio"
	"encoding/binary"
	"io"
	"math"
	"sync"
	"time"
)

// NewTarget creates a Goertzel processor tuned to the given frequency
func NewTarget(freq, sampleRate float64, minDuration time.Duration) *Target {
	t := &Target{
		Frequency:  freq,
		sampleRate: sampleRate,
		blockSize:  optimalBlockSize2(freq, sampleRate, minDuration),
		Threshold:  ToneThreshold,
	}
	t.generateConstants()

	return t
}

// Target is a target frequency detector.  It is a low-level tool which
// implements the Goertzel algorithm to detect the presence of a frequency on a
// block-wise basis.
type Target struct {

	//
	// Constants
	//

	// UseOptimized indicates that an optimized (phase-insensitive) Goertzel should be used for faster arithmetic
	UseOptimized bool

	// Frequency in Hz
	Frequency float64

	// Threshold is the threshold at which this frequency is determined to be present
	Threshold float64

	// sampleRate is the number of times per second that we should receive a sample
	sampleRate float64

	sin       float64
	cos       float64
	coeff     float64
	blockSize int

	//
	//  Working Variables
	//

	q1 float64
	q2 float64

	realM float64
	imagM float64

	// Magnitude2 is the square of the magnitude of the last-processed block
	Magnitude2 float64

	// lastChange records the timestamp of the last change in state
	lastChange time.Time

	// state indicates whether the target frequency is currently detected
	state bool

	// blockReader variables for managing output of block summaries
	blockReaderPresent bool
	blockReader        chan *BlockSummary
	blockReaderMu      sync.Mutex

	stopped bool

	mu sync.Mutex
}

// SetBlockSize overrides automatic calculation of the optimal N (block size) value and uses the one provided instead
func (t *Target) SetBlockSize(n int) {
	t.blockSize = n
	t.generateConstants()
}

func (t *Target) generateConstants() {
	N := float64(t.blockSize)
	rate := t.sampleRate

	k := math.Floor(0.5 + (N*t.Frequency)/rate)
	w := (2.0 * math.Pi / N) * float64(k)
	t.cos = math.Cos(w)
	t.sin = math.Sin(w)
	t.coeff = 2.0 * t.cos
}

// Read processes incoming samples through the Target goertzel
func (t *Target) Read(in io.Reader) error {
	return t.ingest(in)
}

func (t *Target) ingest(in io.Reader) (err error) {
	var i int
	var sample int16
	var q float64

	defer t.Stop()

	r := bufio.NewReader(in)

	t.reset()

	for t.stopped == false {
		var buf = make([]byte, 2)
		buf[0], _ = r.ReadByte()
		buf[1], err = r.ReadByte()
		if err != nil {
			return err
		}

		sample = int16(binary.LittleEndian.Uint16(buf))

		i++
		q = t.coeff*t.q1 - t.q2 + float64(sample)
		t.q2 = t.q1
		t.q1 = q

		if i == t.blockSize {
			t.calculateMagnitude()
			t.sendBlockSummary()
			t.reset()
			i = 0
		}
	}
	return
}

func (t *Target) calculateMagnitude() {
	if t.UseOptimized {
		t.Magnitude2 = t.q1*t.q1 + t.q2*t.q2 - t.q1*t.q2*t.coeff
		return
	}

	var scalingFactor = float64(t.blockSize) / 2.0

	t.realM = (t.q1 - t.q2*t.cos) / scalingFactor
	t.imagM = (t.q2 * t.sin) / scalingFactor
	t.Magnitude2 = t.realM*t.realM + t.imagM*t.imagM
}

func (t *Target) sendBlockSummary() {
	t.blockReaderMu.Lock()
	if t.blockReaderPresent {
		select {
		case t.blockReader <- t.blockSummary():
		default:
		}
	}
	t.blockReaderMu.Unlock()
}

func (t *Target) blockSummary() *BlockSummary {
	return &BlockSummary{
		Magnitude2: t.Magnitude2,
		Frequency:  t.Frequency,
		Duration:   time.Duration(float64(t.blockSize)/t.sampleRate) * time.Second,
		Samples:    t.blockSize,
		Present:    t.Magnitude2 > t.Threshold,
	}
}

func (t *Target) reset() {
	t.q1 = 0
	t.q2 = 0
}

// Stop terminates the Target processing.  It will close the Events channel and stop processing new data.
func (t *Target) Stop() {
	t.blockReaderMu.Lock()
	if t.blockReaderPresent {
		close(t.blockReader)
		t.blockReader = nil
		t.blockReaderPresent = false
	}
	t.blockReaderMu.Unlock()
	t.reset()

	t.stopped = true
}

// Blocks returns a channel over which the summary of each resulting block from
// the Target frequency processor will be returned.  If Blocks() has already
// been called, nil will be returned.
func (t *Target) Blocks() <-chan *BlockSummary {
	t.blockReaderMu.Lock()
	if t.blockReaderPresent {
		t.blockReaderMu.Unlock()
		return nil
	}

	t.blockReaderPresent = true
	t.blockReader = make(chan *BlockSummary, BlockBufferSize)
	t.blockReaderMu.Unlock()

	return t.blockReader
}

// BlockSummary describes the result of a single block of processing for a Target frequency
type BlockSummary struct {

	// Magnitude2 is the square of the relative magnitude of the frequency in this block
	Magnitude2 float64

	// Frequency is the frequency which was being detected
	Frequency float64

	// Duration is the elapsed time which this block represents
	Duration time.Duration

	// Samples is the number of samples this block represents
	Samples int

	// Present indicates whether the frequency was found in the block, as determined by the target's threshold
	Present bool
}
