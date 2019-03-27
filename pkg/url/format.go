package url

import (
	"fmt"
)

func (u *URL) Format(environmentPrefix string) string {
	if u.Protocol == Protocol_Local {
		return u.formatLocal()
	} else if u.Protocol == Protocol_SSH {
		return u.formatSSH()
	} else if u.Protocol == Protocol_Docker {
		return u.formatDocker(environmentPrefix)
	} else if u.Protocol == Protocol_Ipfs {
		return u.formatIpfs()
	} else if u.Protocol == Protocol_MOSH {
		return u.formatMosh()
	}
	panic("unknown URL protocol")
}

func (u *URL) formatLocal() string {
	return u.Path
}
func (u *URL) formatIpfs() string {
	return u.Path
}

func (u *URL) formatMosh() string {

	result := u.Hostname

	if u.Username != "" {
		result = fmt.Sprintf("%s@%s", u.Username, result)
	}

	if u.Port != 0 {
		result = fmt.Sprintf("%s:%d", result, u.Port)
	}

	result = fmt.Sprintf("%s:%s", result, u.Path)

	return result
}

func (u *URL) formatSSH() string {

	result := u.Hostname

	if u.Username != "" {
		result = fmt.Sprintf("%s@%s", u.Username, result)
	}

	if u.Port != 0 {
		result = fmt.Sprintf("%s:%d", result, u.Port)
	}

	result = fmt.Sprintf("%s:%s", result, u.Path)

	return result
}

const invalidDockerURLFormat = "<invalid-docker-url>"

func (u *URL) formatDocker(environmentPrefix string) string {

	result := u.Hostname

	if u.Path == "" {
		return invalidDockerURLFormat
	} else if u.Path[0] == '/' {
		result += u.Path
	} else if u.Path[0] == '~' || isWindowsPath(u.Path) {
		result += fmt.Sprintf("/%s", u.Path)
	} else {
		return invalidDockerURLFormat
	}

	if u.Username != "" {
		result = fmt.Sprintf("%s@%s", u.Username, result)
	}

	result = dockerURLPrefix + result

	if environmentPrefix != "" {
		for _, variable := range DockerEnvironmentVariables {
			result += fmt.Sprintf("%s%s=%s", environmentPrefix, variable, u.Environment[variable])
		}
	}

	return result
}
