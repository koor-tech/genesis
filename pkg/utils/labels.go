package utils

import "fmt"

func Label(k string, v string) string {
	return fmt.Sprintf("%s=%s", k, v)
}
