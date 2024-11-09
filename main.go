package main

import (
	"compass/compass"
)

func main() {
	server := compass.Server{Port: 3000}
	server.Start()
}
