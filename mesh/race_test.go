//go:build race

package mesh

// raceEnabled is true when the tests run under the race detector, which makes
// sync.Pool drop items at random so that allocation counts mean nothing.
const raceEnabled = true
