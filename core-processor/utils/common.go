package utils

// HashString creates a simple hash for unique filenames
func HashString(s string) uint32 {
	var h uint32
	for _, c := range s {
		h = h*31 + uint32(c)
	}
	return h
}
