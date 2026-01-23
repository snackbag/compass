package main

import (
	"fmt"
	"github.com/snackbag/compass/passplate"
)

func main() {
	text := `<h1>Good morning, <$username/>.</h1><p>You are<%if role == "admin"/>an admin<%elif role == "cool"/>pretty cool<%else/>just chill like that<%end/></p>
`
	node, err := passplate.Read(text)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(passplate.Represent(node, 0))
}
