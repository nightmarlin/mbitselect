// Command mbitselect detects the version of the connected microbit, if present,
// and prints the corresponding tinygo target. Otherwise, it uses the fallback
// target. If multiple microbits are connected, the version of the first one
// will be used (this behaviour duplicates that of the `tinygo flash` command).
//
// Usage:
//
//	mbitselect [flags...]
//
// Flags:
//
//	fallback	[microbit|microbit-v2]	the microbit target to fall back to if one cannot be detected.
//	verbose  	[false|true]          	whether to write additional logging output to stderr.
//
// Examples:
//
//	mbitselect
//	mbitselect -fallback=microbit-v2
//	mbitselect -verbose -fallback=microbit-v2
package main

import (
	"errors"
	"flag"
	"fmt"
	"os"

	"go.uber.org/zap"
)

var log = zap.NewNop()

type microbitVersion string

const (
	microbitVersion1 microbitVersion = "microbit"
	microbitVersion2 microbitVersion = "microbit-v2"
)

func (mv microbitVersion) IsValid() bool {
	switch mv {
	case microbitVersion1, microbitVersion2:
		return true
	default:
		return false
	}
}

func (mv microbitVersion) String() string {
	if mv.IsValid() {
		return string(mv)
	}
	return ""
}

var (
	flagFallbackVersion microbitVersion
	flagVerbose         bool
)

func init() {
	flag.StringVar(
		(*string)(&flagFallbackVersion),
		"fallback",
		microbitVersion1.String(),
		`[microbit|microbit-v2] the version of the microbit platform to fall back to if one is not detected, or if multiple versions are detected.`,
	)
	flag.BoolVar(
		&flagVerbose,
		"verbose",
		false,
		`when true, additional logging output will be written`,
	)

	flag.Parse()

	cfg := zap.NewDevelopmentConfig()
	cfg.OutputPaths = []string{"stderr"}
	cfg.ErrorOutputPaths = []string{"stderr"}
	cfg.DisableStacktrace = true

	lv := zap.ErrorLevel
	if flagVerbose {
		lv = zap.InfoLevel
	}
	cfg.Level.SetLevel(lv)

	newLogger, err := cfg.Build()
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "failed to init logger: %s", err.Error())
	} else {
		log = newLogger
	}
}

func main() {
	target := flagFallbackVersion
	if !target.IsValid() {
		log.Error(
			"invalid value for -fallback flag",
			zap.String("value", string(target)),
			zap.String("note", `must be one of "microbit" or "microbit-v2"`),
		)
		os.Exit(1)
	}

	rmv, err := resolveConnectedMicrobitVersion()
	switch {
	case errors.Is(err, errNotDetected):
		log.Info(
			"unable to detect a connected microbit, using fallback",
			zap.String("note", "try un-mounting and re-mounting the device"),
		)
	case errors.Is(err, errInvalidDetails):
		log.Error(
			"the details file could not be parsed",
			zap.String(
				"note",
				`try un-mounting and re-mounting the device, and ensuring its filesystem is correctly mounted`,
			),
		)
		os.Exit(1)
	case errors.Is(err, errUnknownFirmware):
		log.Error(
			"unknown firmware combination",
			zap.Error(err),
			zap.String(
				"note",
				`please report this issue at https://github.com/nightmarlin/mbitselect/issues with a copy of the /MICROBIT/DETAILS.TXT file`,
			),
		)
		os.Exit(1)
	case err != nil:
		log.Error("failed to detect microbit version", zap.Error(err))
		os.Exit(1)
	case !rmv.IsValid():
		log.DPanic(
			"invalid microbit version returned from resolveConnectedMicrobitVersion",
			zap.String("resolved_version", string(rmv)),
			zap.String("note", "report this issue at https://github.com/nightmarlin/mbitselect/issues"),
		)
	default:
		target = rmv
	}

	if _, err := fmt.Fprint(os.Stdout, target); err != nil {
		log.Error("failed to print resolved microbit platform", zap.Error(err))
		os.Exit(1)
	}
}

var (
	errNotDetected     = fmt.Errorf("not detected")
	errInvalidDetails  = fmt.Errorf("invalid details file")
	errUnknownFirmware = fmt.Errorf("unknown firmware combination")
)

func resolveConnectedMicrobitVersion() (microbitVersion, error) {
	path, err := getMBitPathOnCurrentPlatform()
	if err != nil {
		return "", fmt.Errorf("locating microbit: %w", err)
	}

	fd, err := parseDetailsFile(path)
	if err != nil {
		return "", fmt.Errorf("reading firmware details: %w", err)
	}

	mv, ok := firmwareDetailsToMicrobitVersion[fd]
	if !ok {
		return "", fmt.Errorf("%w: %s", errUnknownFirmware, fd)
	}
	return mv, nil
}
