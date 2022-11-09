package profile

import (
	"log"
	"testing"
	"time"
)

func Test11Params(t *testing.T) {
	log.Print("testing 11 parameters...")

	t0 := time.Now()
	stats, err := Get11Stats()
	total := time.Since(t0)
	if err != nil {
		log.Print(err)
	} else {
		log.Printf("took %s", total)
		log.Print(stats)
	}
}

func TestGetCPUAndMemStats(t *testing.T) {
	log.Print("testing cpu and mem stats...")

	stats, err := GetCPUAndMemStats()
	if err != nil {
		log.Print(err)
	} else {
		log.Print(stats)
	}
}
