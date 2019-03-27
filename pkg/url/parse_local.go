package url

func parseLocal(raw string) (*URL, error) {
	return &URL{
		Protocol: Protocol_Local,
		Path:     raw,
	}, nil
}
