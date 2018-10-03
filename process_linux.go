// +build linux

package ps

import (
	"fmt"
	"io/ioutil"
	"log"
	"strconv"
	"strings"
)

// Refresh reloads all the data associated with this process.
func (p *UnixProcess) Refresh() error {
	statPath := fmt.Sprintf("/proc/%d/stat", p.pid)
	dataBytes, err := ioutil.ReadFile(statPath)
	if err != nil {
		return err
	}

	// First, parse out the image name
	data := string(dataBytes)
	binStart := strings.IndexRune(data, '(') + 1
	binEnd := strings.IndexRune(data[binStart:], ')')
	p.binary = data[binStart : binStart+binEnd]

	// Move past the image name and start parsing the rest
	data = data[binStart+binEnd+2:]
	_, err = fmt.Sscanf(data,
		"%c %d %d %d",
		&p.state,
		&p.ppid,
		&p.pgrp,
		&p.sid)

	return err
}

// Returns start time of process, in number of clock ticks after
// system boot. See "man 5 proc" -> /proc/[pid]/stat -> field 22
// for details
func Starttime(pid int) (int, error) {
	if exists, _ := findProcess(pid); exists != nil {
		procStat, err := ioutil.ReadFile("/proc/" + strconv.Itoa(pid) + "/stat")
		if err != nil {
			log.Fatal(err)
		}

		statData := strings.Split(string(procStat), " ")
		startTime, err := strconv.Atoi(statData[21])
		if err != nil {
			return 0, err
		}

		return startTime, nil
	}
	return 0, nil
}
