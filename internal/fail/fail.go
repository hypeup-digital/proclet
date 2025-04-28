package fail

import (
	"fmt"
	"os"
)

func FatalOnErr(err error) {
	if err != nil {
		Fatal(err)
	}
}

func Fatal(i ...interface{}) {
	fmt.Fprintln(os.Stderr, i...)
	os.Exit(1)
}
