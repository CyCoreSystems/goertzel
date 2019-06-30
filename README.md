# goertzel - Golang goertzel tone detection library
[![](https://godoc.org/github.com/CyCoreSystems/goertzel?status.svg)](http://godoc.org/github.com/CyCoreSystems/goertzel)

This library provides tools for tone detection using the goertzel algorithm.
All data is expected to be in 16-bit signed linear format, and there may be
hidden assumptions.  It was built to service telephony-oriented functionality.

Most users will simply make use of the high-level [DetectTone](https://godoc.org/github.com/CyCoreSystems/goertzel#DetectTone) function.  However, lower-level block-wise control is available by directly manipulating the [Target](https://godoc.org/github.com/CyCoreSystems/goertzel#Target) detector.

# Contributing

Contributions welcomed. Changes with tests and descriptive commit messages will get priority handling.  

