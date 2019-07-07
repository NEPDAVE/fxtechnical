package fxtechnical

import (
	"log"
)

var (
	logger *log.Logger
)

func FxTechInit(fxTechLogger *log.Logger) {
	logger = fxTechLogger
}
