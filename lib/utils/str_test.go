package utils_test

import (
	"fmt"

	"github.com/zkep/my-geektime/lib/utils"
)

func ExampleNewStrGenerator() {
	gen := utils.NewStrGenerator(utils.StrGeneratorWithChars(utils.SimpleChars))
	str, err := gen.Encode(1)
	fmt.Println(str, err)
	// Output:
	// 3 <nil>
}
