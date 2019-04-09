package util

func Bold(str string) string {
	return "\033[1m" + str + "\033[0m"
}

func Strikethrough(str string) string {
	return "\033[9m" + str + "\033[0m"
}
