package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/snackbag/compass/v2"
)

func main() {
	data, _ := os.ReadFile("config.json")

	var config compass.ServerConfiguration
	json.Unmarshal(data, &config)

	server := compass.NewServer(compass.FillUnsets(config))

	r, _ := json.MarshalIndent(server.Config, "", "    ")
	fmt.Println(string(r))

	server.MustRun()
}
