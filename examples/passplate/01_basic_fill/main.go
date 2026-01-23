package main

import (
	"fmt"
	"github.com/snackbag/compass/passplate"
)

func main() {
	fmt.Println(passplate.Represent(passplate.Read("<h1>Good morning, <$username/>!</h1>"), -1))
}
