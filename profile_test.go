package profile

import (
	"log"
	"testing"
)

func Test11Params(t *testing.T) {
	log.Print("testing 11 parameters...")

	stats, err := Get11Stats()
	if err != nil {
		log.Print(err)
	} else {
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
