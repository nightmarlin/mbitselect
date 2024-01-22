package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type firmwareDetails struct {
	bootloader string
	iface      string // `interface` is a reserved keyword
}

func (fd firmwareDetails) String() string {
	return fmt.Sprintf("(bootloader: %q; interface: %q)", fd.bootloader, fd.iface)
}

// source: https://tech.microbit.org/software/daplink-interface/#daplink-software
var firmwareDetailsToMicrobitVersion = map[firmwareDetails]microbitVersion{
	{bootloader: "0234", iface: "0234"}: microbitVersion1, // 1.3
	{bootloader: "0234", iface: "0241"}: microbitVersion1, // 1.3b
	{bootloader: "0243", iface: "0249"}: microbitVersion1, // 1.5

	{bootloader: "0255", iface: "0255"}: microbitVersion2, // 2.00
	{bootloader: "0256", iface: "0256"}: microbitVersion2, // 2.20
	{bootloader: "0257", iface: "0257"}: microbitVersion2, // 2.21

	{
		bootloader: "0255",
		iface:      "0258",
	}: microbitVersion2, // 2.00 with bonus firmware (personal mbit has this setup)
}

const (
	mbitDetailsFile         = "DETAILS.TXT"
	headerBootloaderVersion = "Bootloader Version"
	headerIFaceVersion      = "Interface Version"
)

func parseDetailsFile(mbitMountPath string) (firmwareDetails, error) {
	f, err := os.Open(filepath.Join(mbitMountPath, mbitDetailsFile))
	if err != nil {
		return firmwareDetails{}, fmt.Errorf("opening microbit details file: %w", err)
	}
	defer func() { _ = f.Close() }()

	var fd firmwareDetails

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		if scanner.Err() != nil {
			return firmwareDetails{}, fmt.Errorf("reading file: %w", err)
		}
		line := strings.Trim(scanner.Text(), "\r\n")

		if strings.HasPrefix(line, headerBootloaderVersion) {
			sl := strings.Split(line, ": ")
			fd.bootloader = sl[len(sl)-1]
		}
		if strings.HasPrefix(line, headerIFaceVersion) {
			sl := strings.Split(line, ": ")
			fd.iface = sl[len(sl)-1]
		}
	}

	if fd.bootloader == "" || fd.iface == "" {
		return firmwareDetails{}, errInvalidDetails
	}
	return fd, nil
}
