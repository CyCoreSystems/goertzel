package goertzel

import (
	"log"
	"time"
)

var defaultMinimumDuration = 50 // 50ms

// BlockBufferSize is the number of blocks to buffer for slow readers, when sending BlockSummaries from Targets
var BlockBufferSize = 50

const (

	// BlockSizeNorthAmerica is the optimum block size for North America progress tones: 350 440 480 620 850 1400 1800 Hz
	BlockSizeNorthAmerica = 183

	// BlockSizeSouthAmerica is the optimum block size for Costa Rica and Brazil: 425Hz
	BlockSizeSouthAmerica = 188

	// BlockSizeUKDisconnect is the optimum block size for UK disconnect tone: 400Hz
	BlockSizeUKDisconnect = 160

	// BlockSizeDTMF is the optimum block size for DTMF detection
	BlockSizeDTMF = 102

	// RateTelephony is the standard telephony rate of 8kHz
	RateTelephony = 8000.0

	// ToneThreshold is the standard threshold for tone detection
	ToneThreshold = 7.8e7
)

// DTMFFrequencies is the list of frequencies used by standard North American DTMF
var DTMFFrequencies = []float64{697.0, 770.0, 852.0, 941.0, 1209.0, 1336.0, 1477.0, 1633.0}

// NATelephonyFrequencies is the list of common frequencies used in the North American telephony space
var NATelephonyFrequencies = []float64{350.0, 440.0, 480.0, 620.0, 850.0, 1400.0, 1800.0}

// ContactIDFrequencies is the list of signaling frequencies used in SIA ContactID
var ContactIDFrequencies = []float64{1400.0, 2300.0}

// NOTE: this only properly optimizes for integer frequencies (in Hz)
func optimalBlockSize(ω, rate float64, minDuration time.Duration) int {

	// base on 50ms as minimum Duration duration
	// (rate/s)*N = minDuration * s  -> N = minDuration * rate
	// N = 50ms / rate
	// N = (50/1000) / rate
	var maxN = int(minDuration.Seconds() * rate)

	// Find the set of frequency constants (k)
	// ω = A(S/N) -> N = A(S/ω) -> N = Ak -> k = S/ω
	k := rate / ω

	// Find lowest integer j for which Ak is also an integer
	var optimizedK float64
	for i := 1; i < int(maxN); i++ {
		j := float64(i) * k
		if j == float64(int64(j)) {
			optimizedK = j
			break
		}
	}
	if optimizedK == 0 {
		// N = Ak cannot be resolved within maxN, so no further optimization is possible
		//log.Printf("frequency %f cannot be optimized within maxN %d", ω, maxN)
		return int(maxN)
	}

	for n := maxN; n > 1; n-- {
		a := float64(n) / float64(k)
		if a == float64(int64(a)) {
			// a is an integer
			//log.Printf("a (%f) is an integer coefficient of rate(%f)/N(%d)", ω, rate, n)
			return n
		}
	}

	// Return best optimization (most common matches)
	log.Printf("failed to determine better N than max: %d", maxN)
	return maxN
}

func optimalBlockSize2(f, rate float64, minDuration time.Duration) int {
	//var durationSamples int64
	var periodsInBlock int64

	//durationSamples = (int64(minDuration*time.Second) * int64(rate) * 9) / 10

	blockSize := 20 * int64(rate) / 1000

	periodsInBlock = blockSize * int64(f) / int64(rate)
	if periodsInBlock < 5 {
		periodsInBlock = 5
	}

	blockSize = periodsInBlock * int64(rate) / int64(f)

	return int(blockSize)
}

/*
// optimalBlockSizeMultiFreq calculation:
//   - less than half the minimum duration test size for sampling
//   - the largest value of sampleRate/N which evenly divides among each target frequency
//
// NOTE: this only properly optimizes for integer frequencies (in Hz)
func optimalBlockSizeMultiFreq(o *Options) int {

	rate := float64(o.Rate)
	minDet := float64(o.MinDuration)

	// base on 50ms as minimum Duration duration
	// (rate/s)*N = minDuration * s  -> N = minDuration * rate
	// N = 50ms / rate
	// N = (50/1000) / rate
	var maxN = (minDet * rate) / 1000.0
	//log.Println("maxN:", maxN)

	// Start with bestN = maxN
	var bestN = struct {
		// Value of N
		N int

		// Number of frequencies matched
		Matches int
	}{
		N:       int(maxN),
		Matches: 1,
	}

	// Find the set of frequency constants (k)
	// ω = A(S/N) -> N = A(S/ω) -> N = Ak -> k = S/ω
	var constants []float64
	for _, f := range o.Frequencies {
		k := float64(o.Rate) / f
		//log.Printf("k for %f is %f", f, k)

		// Find lowest integer A for which Ak is also an integer
		var optimizedK float64
		for i := 1; i < int(maxN); i++ {
			j := float64(i) * k
			if j == float64(int64(j)) {
				optimizedK = j
				break
			}
		}
		if optimizedK == 0 {
			// N = Ak cannot be resolved within maxN, so no further optimization is possible
			//log.Printf("frequency %f cannot be optimized within maxN %f", f, maxN)
			return int(maxN)
		}

		constants = append(constants, optimizedK)
	}

	for n := maxN; n > 1; n-- {
		var matches int
		for _, k := range constants {
			a := float64(n) / float64(k)
			if a == float64(int64(a)) {
				// a is an integer
				//log.Printf("a (%f) is an integer coefficient of rate(%d)/N(%f)", a, o.Rate, n)
				matches++
			}
		}

		if matches == len(o.Frequencies) {
			// Best possible N found
			//log.Println("best possible N:", n)
			return int(n)
		}

		if matches > bestN.Matches {
			bestN.Matches = matches
			bestN.N = int(n)
		}
	}

	// Return best optimization (most common matches)
	return bestN.N
}
*/
