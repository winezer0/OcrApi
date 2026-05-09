package utils

import (
	"encoding/base64"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"
)

func NormalizeBase64(text string) string {
	text = strings.TrimSpace(text)
	if text == "" {
		return ""
	}
	if strings.Contains(text, ",") && strings.HasPrefix(strings.ToLower(text), "data:image") {
		parts := strings.SplitN(text, ",", 2)
		if len(parts) == 2 {
			text = parts[1]
		}
	}
	return strings.TrimSpace(text)
}

func ImageToBase64(imagePath string) (string, error) {
	data, err := os.ReadFile(imagePath)
	if err != nil {
		return "", fmt.Errorf("读取图片文件失败: %w", err)
	}
	return base64.StdEncoding.EncodeToString(data), nil
}

func ExtractResultNumber(text string) string {
	if text == "" {
		return ""
	}

	expRegex := regexp.MustCompile(`(-?\d+)\s*([+\-xX*/÷])\s*(-?\d+)`)
	expMatch := expRegex.FindStringSubmatch(text)

	if len(expMatch) == 4 {
		left, _ := strconv.Atoi(expMatch[1])
		operator := expMatch[2]
		right, _ := strconv.Atoi(expMatch[3])

		switch operator {
		case "+":
			return strconv.Itoa(left + right)
		case "-":
			return strconv.Itoa(left - right)
		case "x", "X", "*":
			return strconv.Itoa(left * right)
		case "÷", "/":
			if right != 0 {
				return strconv.Itoa(left / right)
			}
		}
	}

	numRegex := regexp.MustCompile(`-?\d+`)
	numbers := numRegex.FindAllString(text, -1)
	if len(numbers) > 0 {
		return numbers[len(numbers)-1]
	}

	return ""
}
