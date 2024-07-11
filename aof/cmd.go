package aof

import (
	"strconv"
)

func buildCmdLine(cmdName string, args ...[]byte) [][]byte {
	result := make([][]byte, 1)
	result[0] = []byte(cmdName)
	for _, arg := range args {
		result = append(result, arg)
	}
	return result
}

func SetCmd(args ...[]byte) [][]byte {
	return buildCmdLine("SET", args...)
}

func ExpireAtCmd(key string, ttl int64) [][]byte {
	return buildCmdLine("EXPIRE", []byte(key), []byte(strconv.Itoa(int(ttl))))
}

func ZAddCmd(args ...[]byte) [][]byte {
	return buildCmdLine("ZADD", args...)
}
