package main

import (
    "strings"
    "strconv"
)

func substringInSlice(str string, list []string) bool {
	for _, v := range list {
		if strings.Contains(str, v) {
			return true
		}
	}
	return false
}

func decodeSuffix(str string) int64 {
    var mul byte
    var num float64

    mulindex := strings.IndexAny(str, "KMGT")
    if (mulindex != -1) {
        mul = str[mulindex]
        str = strings.TrimRight(str, "KMGT")
    }
    num, _ = strconv.ParseFloat(str, 64)
    switch string(mul) {
    case "K":
        return int64(num * 1000)
    case "M":
        return int64(num * 1000 * 1000)
    case "G":
        return int64(num * 1000 * 1000 * 1000)
    case "T":
        return int64(num * 1000 * 1000 * 1000 * 1000)
    default:
        return int64(num)
    }
}
