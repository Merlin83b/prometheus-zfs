package main

import (
	"errors"
	"fmt"
	"log"
    "os"
	"os/exec"
	"strconv"
	"strings"
)

type zpool struct {
	name     string
	capacity int64
	healthy  bool
	status   string
	online   int64
	faulted  int64
    iops_r   int64
    iops_w   int64
    bandwidth_r int64
    bandwidth_w int64
}

func (z *zpool) checkHealth(output string) (err error) {
	output = strings.Trim(output, "\r\n")
	if output == "ONLINE" {
		z.healthy = true
	} else if output == "DEGRADED" || output == "FAULTED" {
		z.healthy = false
	} else {
		z.healthy = false // just to make sure
		err = errors.New("Unknown status")
	}
	return err
}

func (z *zpool) getCapacity(output string) (err error) {
	s := strings.Split(output, "%")[0]
	z.capacity, err = strconv.ParseInt(s, 0, 8)
	if err != nil {
		return err
	}
	return err
}

func (z *zpool) getProviders(output string) (err error) {
	nonProviderLines := []string{
		z.name,
		"state:",
        "scan:",
        "config:",
		"mirror-",
		"raid0-",
		"raid10-",
		"raidz-",
		"raidz2-",
		"raidz3-",
        "errors",
    }
	lines := strings.Split(output, "\n")
    for index, line := range lines {
       lines[index] = strings.TrimRight(line, "\r\n")
    }
	z.status = strings.Split(lines[1], " ")[2]

	// Count all providers, ONLINE and FAULTED
	var fcount int64
	var dcount int64
	for _, line := range lines {
		if (strings.Contains(line, "FAULTED") || strings.Contains(line, "UNAVAIL")) && !substringInSlice(line, nonProviderLines) {
			fcount = fcount + 1
		} else if strings.Contains(line, "ONLINE") && !substringInSlice(line, nonProviderLines) {
			dcount = dcount + 1
		}
	}
	z.faulted = fcount
	z.online = dcount

	if z.status != "ONLINE" && z.status != "DEGRADED" && z.status != "FAULTED" {
		z.faulted = 1 // fake faulted if there is a parsing error or other status
		err = errors.New("Error parsing faulted/unavailable providers")
	}
	return
}

func (z *zpool) getIostat(output string) (err error) {
    lines := strings.Split(output, "\n")
    for index, line := range lines {
       lines[index] = strings.TrimRight(line, "\r\n")
    }
    values := strings.Fields(lines[len(lines)-2])
    z.iops_r = decodeSuffix(values[3])
    z.iops_w = decodeSuffix(values[4])
    z.bandwidth_r = decodeSuffix(values[5])
    z.bandwidth_w = decodeSuffix(values[6])
    return
}

func (z *zpool) getStatus() {
	output := runZpoolCommand([]string{"status", z.name})
	err := z.getProviders(output)
	if err != nil {
		log.Fatal("Error parsing zpool status")
	}
	output = runZpoolCommand([]string{"list", "-H", "-o", "health", z.name})
	err = z.checkHealth(output)
	if err != nil {
		log.Fatal("Error parsing zpool list -H -o health ", z.name)
	}
	output = runZpoolCommand([]string{"list", "-H", "-o", "cap", z.name})
	err = z.getCapacity(output)
	if err != nil {
		log.Fatal("Error parsing zpool capacity")
	}
    output = runZpoolCommand([]string{"iostat", z.name, "1", "2"})
    err = z.getIostat(output)
    if err != nil {
        log.Fatal("Error parsing zpool iostat")
    }
}

func checkExistance(pool string) (err error) {
	output := runZpoolCommand([]string{"list", pool})
	if strings.Contains(fmt.Sprintf("%s", output), "no such pool") {
		err = errors.New("No such pool")
	}
	return
}

func runZpoolCommand(args []string) string {
	zpoolPath, err := exec.LookPath("zpool")
	if err != nil {
		log.Fatal("Could not find zpool in PATH")
	}
	cmd := exec.Command(zpoolPath, args...)
    cmd.Stdin = os.Stdin
	out, _ := cmd.CombinedOutput()
	return fmt.Sprintf("%s", out)
}
