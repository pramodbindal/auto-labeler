package main

import (
	"github.com/pramodbindal/auto-labeler/cmd/controller/labeler"
	"knative.dev/pkg/injection/sharedmain"
)

func main() {
	sharedmain.Main("auto-labeler-controller", labeler.NewController)
}
