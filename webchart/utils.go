package webchart

import "fmt"

func HashCode(text string) int {
	result := 0
	for _, symbol := range []byte(text) {
		result = int(symbol) + ((result << 5) - result)
	}

	return result
}

func IntToHexColor(i int) string {
	result := fmt.Sprintf("#%0.2X", i)
	bound := len(result)
	if len(result) > 7 {
		bound = 7
	}
	return result[:bound]
}

func ColorFromString(text string) string {
	return IntToHexColor(HashCode(text))
}
