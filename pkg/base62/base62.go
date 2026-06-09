// Package base62 содержить функции (де)кодирования с помощью алгоритма base62
package base62

import (
	"errors"
	"math"
	"strings"
)

// Алфавит Base62: 0-9, a-z, A-Z (порядок важен для сохранения сортировки в БД)
const (
	chars           = "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	base            = 62
	miniLettersDiff = 10
	MaxLettersDiff  = 36
)

// Encode преобразует числовой ID в короткую строку Base62
func Encode(id int64) string {
	if id == 0 {
		return string(chars[0])
	}

	var sb strings.Builder
	for id > 0 {
		remainder := id % base
		sb.WriteByte(chars[remainder])
		id /= base
	}

	// Переворачиваем строку, так как остатки получаются в обратном порядке
	encoded := sb.String()
	runes := []rune(encoded)
	for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
		runes[i], runes[j] = runes[j], runes[i]
	}
	return string(runes)
}

// Decode преобразует короткую строку Base62 обратно в числовой ID
func Decode(encoded string) (int64, error) {
	var id int64
	length := len(encoded)

	for i, char := range encoded {
		var power = int64(length - 1 - i)
		var value int

		// Быстрое определение значения символа
		switch {
		case char >= '0' && char <= '9':
			value = int(char - '0')
		case char >= 'a' && char <= 'z':
			value = int(char - 'a' + miniLettersDiff)
		case char >= 'A' && char <= 'Z':
			value = int(char - 'A' + MaxLettersDiff)
		default:
			return 0, errors.New("недопустимый символ в строке Base62")
		}

		id += int64(value) * int64(math.Pow(base, float64(power)))
	}

	return id, nil
}
