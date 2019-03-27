package url

func isWindowsPath(raw string) bool {
	return len(raw) >= 3 &&
		((raw[0] >= 'a' && raw[0] <= 'z') || (raw[0] >= 'A' && raw[0] <= 'Z')) &&
		raw[1] == ':' &&
		(raw[2] == '\\' || raw[2] == '/')
}
