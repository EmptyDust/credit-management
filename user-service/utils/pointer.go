package utils

// DerefString returns the value of a *string or "" if nil
func DerefString(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}
