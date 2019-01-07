package internal

import (
	"fmt"
	"os"
)

func DieOnError(err error) {
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
