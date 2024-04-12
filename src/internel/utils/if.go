package utils

func TIF[T any](t bool, a1 T, a2 T) T {
	if t {
		return a1
	}
	return a2
}
