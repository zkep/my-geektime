package utils

import (
	"fmt"
	"math"
	"math/rand"
	"strconv"
	"strings"
	"time"
)

const (
	DefaultChars = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"
	SimpleChars  = "23456789ABCDEFGHKMNOPQRSTUVWXYZ"
)

func init() {
	rand.NewSource(time.Now().UnixNano())
}

type StrGenerator struct {
	chars string
}

type StrGeneratorFunc func(*StrGenerator)

func StrGeneratorWithChars(chars string) StrGeneratorFunc {
	return func(c *StrGenerator) {
		c.chars = chars
	}
}

func NewStrGenerator(opts ...StrGeneratorFunc) *StrGenerator {
	c := &StrGenerator{chars: DefaultChars}
	for _, opt := range opts {
		opt(c)
	}
	return c
}

func (c *StrGenerator) Random(n int) string {
	bytes := make([]byte, 0, n)
	for i := 0; i < n; i++ {
		bytes = append(bytes, c.chars[rand.Intn(len(c.chars))])
	}
	return string(bytes)
}

func (c *StrGenerator) EncodeWithSeq(seq int64) (string, error) {
	basestep := time.Now().UnixNano() / 1e5
	b := []byte(fmt.Sprintf("%d", basestep+seq))
	ReverseByte(b)
	num, err := strconv.ParseInt(string(b), 10, 64)
	if err != nil {
		return "", err
	}
	return c.Encode(num)
}

func (c *StrGenerator) Encode(num int64) (string, error) {
	bytes := []byte{}
	if num > 60 {
		bytes = []byte(fmt.Sprintf("%X", num))
	} else {
		for num > 0 {
			bytes = append(bytes, c.chars[num%62])
			num = num / 62
		}
	}
	return string(bytes), nil
}

func (c *StrGenerator) Decode(str string) int64 {
	var num int64
	n := len(str)
	for i := 0; i < n; i++ {
		pos := strings.IndexByte(c.chars, str[i])
		num += int64(math.Pow(62, float64(n-i-1)) * float64(pos))
	}
	return num
}

func ReverseByte(a []byte) {
	for left, right := 0, len(a)-1; left < right; left, right = left+1, right-1 {
		a[left], a[right] = a[right], a[left]
	}
}
