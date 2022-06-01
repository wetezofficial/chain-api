package utils

func Find(str string, strList []string) int {
	for i, item := range strList {
		if item == str {
			return i
		}
	}
	return -1
}

func In(str string, strList []string) bool {
	return Find(str, strList) != -1
}
