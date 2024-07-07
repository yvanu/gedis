package aof

func buildCmdLine(cmdName string, args ...[]byte) [][]byte {
	result := make([][]byte, len(args)+1)
	result[0] = []byte(cmdName)
	for _, arg := range args {
		result = append(result, arg)
	}
	return result
}

func SetCmd(args ...[]byte) [][]byte {
	return buildCmdLine("SET", args...)
}
