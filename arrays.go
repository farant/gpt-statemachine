package main

func Deduplicate(strings []string) []string {
	unique := make(map[string]bool)
	var result []string
	for _, entry := range strings {
		if _, value := unique[entry]; !value {
			unique[entry] = true
			result = append(result, entry)
		}
	}
	return result
}
