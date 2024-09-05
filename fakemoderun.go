package main

import (
	"flag"
	"log"
	"os"
	"os/exec"
	"reflect"
	"strings"
	"syscall"
)

var (
	f_cache = flag.Bool("cache", true, "Prefer Cache Cores")
	f_freq  = flag.Bool("freq", false, "Prefer Frequency Cores")
)

func killProcess(cmd *exec.Cmd) {
	if nil == cmd {
		return
	}
	log.Printf("Terminating Process %s", cmd)
	cmd.Process.Kill()
}

func getMask7950X3D() (uintptr, uintptr, error) {

	// Cores 1-8, 17-24
	cache := 0
	cache |= 1 << 0
	cache |= 1 << 1
	cache |= 1 << 2
	cache |= 1 << 3
	cache |= 1 << 4
	cache |= 1 << 5
	cache |= 1 << 6
	cache |= 1 << 7
	cache |= 1 << 16
	cache |= 1 << 17
	cache |= 1 << 18
	cache |= 1 << 19
	cache |= 1 << 20
	cache |= 1 << 21
	cache |= 1 << 22
	cache |= 1 << 23

	// Cores 9-16, 25-32
	freq := 0
	freq |= 1 << 8
	freq |= 1 << 9
	freq |= 1 << 10
	freq |= 1 << 11
	freq |= 1 << 12
	freq |= 1 << 13
	freq |= 1 << 14
	freq |= 1 << 15
	freq |= 1 << 24
	freq |= 1 << 25
	freq |= 1 << 26
	freq |= 1 << 27
	freq |= 1 << 28
	freq |= 1 << 39
	freq |= 1 << 30
	freq |= 1 << 31

	return uintptr(cache), uintptr(freq), nil
}

func getMask(useCache bool, useFreq bool) (uintptr, error) {
	// HACK: Just Testing for now
	var mask uintptr = 0

	cacheMask, freqMask, err := getMask7950X3D()
	if err != nil {
		return mask, err
	}

	if !useCache && !useFreq {
		return 0xFFFFFFFF, nil
	}

	if useCache {
		mask |= cacheMask
	}

	if useFreq {
		mask |= freqMask
	}

	return mask, nil
}

func main() {
	name := strings.Join(os.Args[1:2], " ")
	args := strings.Join(os.Args[2:], " ")

	mask, err := getMask(*f_cache, *f_freq)
	if err != nil {
		log.Fatalf("Failed to get process mask: %s", err)
	}

	log.Printf("Using CPU Mask: %032b", mask)

	cmd := exec.Command(name, args)
	log.Printf("Starting Process: %s", cmd)
	err = cmd.Start()
	if err != nil {
		log.Fatalf("Failed to Start Process: %s", err)
	}

	defer killProcess(cmd)

	kernel32 := syscall.NewLazyDLL("kernel32.dll")
	setProcessAffinityMask := kernel32.NewProc("SetProcessAffinityMask")
	handle := uintptr(reflect.ValueOf(cmd.Process).Elem().FieldByName("handle").Uint())
	setProcessAffinityMask.Call(uintptr(handle), mask)

	err = cmd.Wait()
	if err != nil {
		log.Fatalf("Process Returned with error: %s", err)
	}

	log.Printf("Process exited successfully")
}
