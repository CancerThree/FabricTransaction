package common

import "strings"

func IsEmptyStr(str string) bool {
	return strings.Trim(str, " ") == ""
}
