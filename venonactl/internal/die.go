package internal

import (
	"os"

	"github.com/sirupsen/logrus"
)

func DieOnError(err error) {
	if err != nil {
		logrus.Error(err)
		os.Exit(1)
	}
}
