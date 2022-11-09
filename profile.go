package profile

import (
	"bytes"
	"errors"
	"log"
	"os/exec"
	"strconv"
	"strings"
	"time"

	cpu "github.com/shirou/gopsutil/v3/cpu"
	mem "github.com/shirou/gopsutil/v3/mem"
	proc "github.com/shirou/gopsutil/v3/process"
)

type DNN_params struct {
	MachineID    string
	Freq         float64
	UserTime     float64
	VMem         float64
	CPUPercent   float64
	Syscalls     float64
	Shared       uint64
	Interrupts   uint64
	SWInterrupts uint64
	PIDs         uint64
	Instructions float64
	MissRatio    float64
	Timestamp    string
}

func GetCPUAndMemStats() (map[string]interface{}, error) {
	freq, err := cpu_freq()
	if err != nil {
		return nil, err
	}
	mem_stats, err := mem_stats()
	if err != nil {
		return nil, err
	}
	percent, err := cpu_percent()
	if err != nil {
		return nil, err
	}
	stats := map[string]interface{}{
		"freq":       freq,
		"vmem":       mem_stats["vmem"],
		"cpupercent": percent,
		"shared":     mem_stats["shared"],
		"timestamp":  time.Now().Format("15:04:05"),
	}
	return stats, nil
}

func Get11Stats() (map[string]interface{}, error) {

	percentQuery := make(chan float64)
	go cpu_percent_async(percentQuery)

	cacheStatsQuery := make(chan map[string]float64)
	go cache_stats_async(cacheStatsQuery)

	freq, err := cpu_freq()
	if err != nil {
		return nil, err
	}

	user_time, err := user_time()
	if err != nil {
		return nil, err
	}

	mem_stats, err := mem_stats()
	if err != nil {
		return nil, err
	}

	interrupts, err := interrupts()
	if err != nil {
		return nil, err
	}

	sw_interrupts, err := sw_interrupts()
	if err != nil {
		return nil, err
	}

	pids, err := pids()
	if err != nil {
		return nil, err
	}

	percent := <-percentQuery
	close(percentQuery)
	if percent == 0.0 {
		return nil, errors.New("no cpu percent")
	}

	cacheStats := <-cacheStatsQuery
	close(cacheStatsQuery)
	if cacheStats == nil {
		return nil, errors.New("no cache stats")
	}

	stats := map[string]interface{}{
		"freq":         freq,
		"userTime":     user_time,
		"vmem":         mem_stats["vmem"],
		"cpupercent":   percent,
		"syscalls":     sw_interrupts,
		"shared":       mem_stats["shared"],
		"interrupts":   interrupts,
		"swinterrupts": sw_interrupts,
		"pids":         pids,
		"instructions": cacheStats["instructions"],
		"missRatio":    cacheStats["missRatio"],
		"timestamp":    time.Now().Format("15:04:05"),
	}
	return stats, nil
}

// Python equivalent: ps.cpu_percent()
func cpu_percent() (float64, error) {
	A, err := cpu.Percent(time.Duration(time.Second), false)
	if err != nil {
		return 0.0, err
	}

	return A[0], nil
}

func cpu_percent_async(percent chan float64) {
	log.Print("running cpu percent")
	A, err := cpu.Percent(time.Duration(time.Second), false)
	if err != nil {
		log.Print(err)
		percent <- 0.0
		return
	}
	log.Print("done cpu percent", A[0])
	percent <- A[0]
}

// Python equivalent: ps.cpu_freq()
func cpu_freq() (float64, error) {
	grep := exec.Command("grep", "CPU MHz")
	lscpu := exec.Command("lscpu")

	pipe, _ := lscpu.StdoutPipe()
	defer pipe.Close()
	grep.Stdin = pipe

	lscpu.Start()
	res, _ := grep.Output()
	freq := strings.Split(string(res), ":")[1]
	freq_ := strings.ReplaceAll(freq, " ", "")
	freq__ := strings.ReplaceAll(freq_, "\n", "")

	stat, err := strconv.ParseFloat(freq__, 32)
	if err != nil {
		return 0.0, err
	}
	return stat, nil
}

// Python equivalent: ps.cpu_times()[0]
func user_time() (float64, error) {
	times, err := cpu.Times(false)
	if err != nil {
		return 0.0, err
	}

	return times[0].User, nil
}

// Python equivalent: ps.virtual_memory()[2]
func virt_mem() (float64, error) {
	stat, err := mem.VirtualMemory()
	if err != nil {
		return 0.0, err
	}
	return stat.UsedPercent, nil
}

// Python equivalent: ps.virtual_memory()[9]
func shared_mem() (uint64, error) {
	stat, err := mem.VirtualMemory()
	if err != nil {
		return 0.0, err
	}
	return stat.Shared, nil
}

func mem_stats() (map[string]interface{}, error) {
	stat, err := mem.VirtualMemory()
	if err != nil {
		return nil, err
	}
	return map[string]interface{}{"vmem": stat.UsedPercent, "shared": stat.Shared}, nil
}

// Python equivalent: ps.cpu_stats()[1]
func interrupts() (int, error) {
	grepIntr := exec.Command("grep", "intr")
	cat := exec.Command("cat", "/proc/stat")

	pipe, _ := cat.StdoutPipe()
	grepIntr.Stdin = pipe
	defer pipe.Close()

	cat.Start()
	resIntr, _ := grepIntr.Output()

	intr, err := interruptHelper(resIntr)
	if err != nil {
		return 0, err
	}

	return intr, nil
}

// Python equivalent: ps.cpu_stats()[2]
func sw_interrupts() (int, error) {
	grepSwintr := exec.Command("grep", "softirq")
	cat := exec.Command("cat", "/proc/stat")

	pipe, _ := cat.StdoutPipe()
	grepSwintr.Stdin = pipe
	defer pipe.Close()

	cat.Start()
	resSwintr, _ := grepSwintr.Output()

	swintr, err := interruptHelper(resSwintr)
	if err != nil {
		return 0, err
	}
	return swintr, nil
}

func interruptHelper(output []byte) (int, error) {
	some_ := strings.Split(string(output), " ")[1]
	str_val := some_
	val, err := strconv.Atoi(str_val)
	if err != nil {
		return 0, err
	}
	return val, nil
}

// Python equivalent: len(ps.pids())
func pids() (int, error) {
	pids, err := proc.Pids()
	if err != nil {
		return 0, err
	}
	return len(pids), nil
}

// Python equivalent: N/A
func cache_stats() (map[string]float64, error) {
	perf := exec.Command("perf", "stat", "--time", "1000", "-e", "cache-misses,cache-references,instructions")
	var outb, errb bytes.Buffer
	perf.Stdout = &outb
	perf.Stderr = &errb
	err := perf.Run()
	if err != nil {
		return nil, errors.New(errb.String())
	}
	lines := strings.Split(errb.String(), "\n")
	cacheMisses, _ := strconv.ParseFloat(stringClean(strings.Fields(lines[3])[0]), 64)
	cacheRefs, _ := strconv.ParseFloat(stringClean(strings.Fields(lines[4])[0]), 64)
	instructions, _ := strconv.ParseFloat(stringClean(strings.Fields(lines[5])[0]), 64)

	mr := cacheMisses / cacheRefs
	return map[string]float64{"instructions": instructions, "missRatio": mr}, nil
}

func cache_stats_async(cacheStats chan map[string]float64) {
	perf := exec.Command("perf", "stat", "--time", "1000", "-e", "cache-misses,cache-references,instructions")
	var outb, errb bytes.Buffer
	perf.Stdout = &outb
	perf.Stderr = &errb
	log.Print("Running cache stats")
	err := perf.Run()
	if err != nil {
		log.Print(errb.String())
		cacheStats <- nil
		return
	}
	lines := strings.Split(errb.String(), "\n")
	cacheMisses, _ := strconv.ParseFloat(stringClean(strings.Fields(lines[3])[0]), 64)
	cacheRefs, _ := strconv.ParseFloat(stringClean(strings.Fields(lines[4])[0]), 64)
	instructions, _ := strconv.ParseFloat(stringClean(strings.Fields(lines[5])[0]), 64)
	mr := cacheMisses / cacheRefs
	stats := map[string]float64{"instructions": instructions, "missRatio": mr}

	log.Print("done cache stats.", stats)
	cacheStats <- stats
}

func stringClean(s string) string {
	s_ := strings.ReplaceAll(s, " ", "")
	s__ := strings.ReplaceAll(s_, "\n", "")
	s___ := strings.ReplaceAll(s__, "\\n", "")
	s____ := strings.ReplaceAll(s___, ",", "")
	return s____
}
