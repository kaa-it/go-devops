package atargs

import (
	"fmt"
	"os"
)

func test(x int, f func(int)) {
	fmt.Println(x)
	f(x)
}

func main() {
	test(15, os.Exit)
}
