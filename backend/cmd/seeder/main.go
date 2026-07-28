package main

import (
	"os"

	compositionseeder "github.com/MateWorkspace/mate-things/backend/internal/composition/seeder"
)

func main() {
	os.Exit(compositionseeder.Launch())
}
