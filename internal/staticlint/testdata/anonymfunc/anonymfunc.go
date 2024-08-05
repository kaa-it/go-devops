package anonymfunc

import "os"

func main() {
	x := func() {
		os.Exit(1)
	}

	x()
}
