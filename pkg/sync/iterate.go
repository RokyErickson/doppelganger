package sync

func nameUnion(contentMaps ...map[string]*Entry) map[string]bool {
	capacity := 0
	if len(contentMaps) > 0 {
		capacity = len(contentMaps[0])
	}
	result := make(map[string]bool, capacity)

	for _, contents := range contentMaps {
		for name := range contents {
			result[name] = true
		}
	}

	return result
}
