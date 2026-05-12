package utils

func FlattenMap[K comparable, V any](m map[K][]V) []V {
	// 1. Calculate total size to perform only ONE memory allocation
	totalSize := 0
	for _, slice := range m {
		totalSize += len(slice)
	}

	// 2. Initialize slice with the exact required capacity
	result := make([]V, 0, totalSize)

	// 3. Use variadic append (...) to add entire slices at once
	for _, slice := range m {
		result = append(result, slice...)
	}

	return result
}
