package utils

import (
	"log"
	"testing"
)

func TestMathUtils(t *testing.T) {

	t.Run("ConvertFloatToDecimal", func(t *testing.T) {

		val := ConvertFloatToDecimal(278.9879992399)
		log.Println(val)
	})
}
