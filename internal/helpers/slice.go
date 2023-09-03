package helpers

func Contains[T comparable](arr []T, element T) bool {
	for _, elem := range arr {
		if element == elem {
			return true
		}
	}
	return false
}

func Last[T any](arr []T) T {
	if len(arr) != 0 {
		return arr[len(arr)-1]
	} else {
		var ret T
		return ret
	}
}
