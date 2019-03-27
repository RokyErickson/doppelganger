package agent

import (
	"strings"
	//"unicode/utf8"

	"github.com/pkg/errors"

	"github.com/RokyErickson/doppelganger/pkg/prompt"
)

var unameSToGOOS = map[string]string{
	"Linux":     "linux",
	"Darwin":    "darwin",
	"FreeBSD":   "freebsd",
	"NetBSD":    "netbsd",
	"OpenBSD":   "openbsd",
	"DragonFly": "dragonfly",
	"SunOS":     "solaris",
	"Plan9":     "plan9",
}

func unameSIsWindowsPosix(value string) bool {
	return strings.HasPrefix(value, "CYGWIN") ||
		strings.HasPrefix(value, "MINGW") ||
		strings.HasPrefix(value, "MSYS")
}

var unameMToGOARCH = map[string]string{
	"i386":     "386",
	"i486":     "386",
	"i586":     "386",
	"i686":     "386",
	"x86_64":   "amd64",
	"amd64":    "amd64",
	"armv5l":   "arm",
	"armv6l":   "arm",
	"armv7l":   "arm",
	"armv8l":   "arm64",
	"aarch64":  "arm64",
	"mips":     "mips",
	"mipsel":   "mipsle",
	"mips64":   "mips64",
	"mips64el": "mips64le",
	"ppc64":    "ppc64",
	"ppc64le":  "ppc64le",
	"s390x":    "s390x",
}

var osEnvToGOOS = map[string]string{
	"Windows_NT": "windows",
}

var processorArchitectureEnvToGOARCH = map[string]string{
	"x86":   "386",
	"AMD64": "amd64",
}

func probePOSIX(transport Transport) (string, string, error) {

	unameSMBytes := output(transport, "uname -s -m")

	unameSM := strings.Split(strings.TrimSpace(string(unameSMBytes)), " ")
	if len(unameSM) != 2 {
		return "", "", errors.New("invalid uname output")
	}
	unameS := unameSM[0]
	unameM := unameSM[1]

	var goos string
	if unameSIsWindowsPosix(unameS) {
		goos = "windows"
	} else if g, ok := unameSToGOOS[unameS]; ok {
		goos = g
	} else {
		return "", "", errors.New("unknown platform")
	}

	goarch, ok := unameMToGOARCH[unameM]
	if !ok {
		return "", "", errors.New("unknown architecture")
	}

	return goos, goarch, nil
}

func probeWindows(transport Transport) (string, string, error) {
	outputBytes := output(transport, "cmd /c set")
	output := string(outputBytes)
	output = strings.Replace(output, "\r\n", "\n", -1)
	output = strings.TrimSpace(output)
	environment := strings.Split(output, "\n")

	var os, processorArchitecture string
	for _, e := range environment {
		if strings.HasPrefix(e, "OS=") {
			os = e[3:]
		} else if strings.HasPrefix(e, "PROCESSOR_ARCHITECTURE=") {
			processorArchitecture = e[23:]
		}
	}

	goos, ok := osEnvToGOOS[os]
	if !ok {
		return "", "", errors.New("unknown platform")
	}

	goarch, ok := processorArchitectureEnvToGOARCH[processorArchitecture]
	if !ok {
		return "", "", errors.New("unknown architecture")
	}

	return goos, goarch, nil
}

func probe(transport Transport, prompter string) (string, string, bool, error) {

	if err := prompt.Message(prompter, "Probing endpoint (POSIX)..."); err != nil {
		return "", "", false, errors.Wrap(err, "unable to message prompter")
	}
	if goos, goarch, err := probePOSIX(transport); err == nil {
		return goos, goarch, true, nil
	}

	if err := prompt.Message(prompter, "Probing endpoint (Windows)..."); err != nil {
		return "", "", false, errors.Wrap(err, "unable to message prompter")
	}
	if goos, goarch, err := probeWindows(transport); err == nil {
		return goos, goarch, false, nil
	}

	return "", "", false, errors.New("exhausted probing methods")
}
