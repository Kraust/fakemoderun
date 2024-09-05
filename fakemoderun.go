package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"reflect"
	"strconv"
	"strings"
	"syscall"
)

var (
	f_cores = flag.String("cores", "", "List of cores to set the affinitize the process to")
)

func killProcess(cmd *exec.Cmd) {
	if nil == cmd {
		return
	}
	log.Printf("Terminating Process %s", cmd)
	cmd.Process.Kill()
}

func getMask(coreMask string) (uintptr, error) {
	// HACK: Just Testing for now
	var mask uintptr = 0

	ok := false
	coreRanges := strings.Split(coreMask, ",")

	for _, coreRange := range coreRanges {
		cores := strings.Split(coreRange, "-")
		switch len(cores) {
		case 1:
			num, err := strconv.Atoi(cores[0])
			if err != nil {
				return mask, err
			}
			mask |= 1 << num
			ok = true
			break
		case 2:
			min, err := strconv.Atoi(cores[0])
			if err != nil {
				return mask, err
			}

			max, err := strconv.Atoi(cores[1])
			if err != nil {
				return mask, err
			}

			for idx := min; idx <= max; idx++ {
				mask |= 1 << idx
				ok = true
			}
			break
		default:
			return mask, fmt.Errorf("Failed to parse core range [%s]", cores)
		}
	}

	if !ok {
		log.Printf("No cores specified. Allocating all cores.")
		mask = 0xffffffff
	}

	return mask, nil
}

func main() {
	flag.Parse()

	name := strings.Join(flag.Args()[:1], " ")
	args := strings.Join(flag.Args()[1:], " ")

	log.Printf("Using core mask [%s]", *f_cores)

	mask, err := getMask(*f_cores)
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

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	go func(cmd *exec.Cmd) {
		<-sigChan
		killProcess(cmd)
		os.Exit(1)
	}(cmd)

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
