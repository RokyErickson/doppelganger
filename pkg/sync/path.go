package sync

func pathJoin(base, leaf string) string {
	if base == "" {
		return leaf
	}
	return base + "/" + leaf
}
