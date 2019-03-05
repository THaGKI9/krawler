package main

import (
	"os"

	"github.com/thagki9/krawler"
)

func main() {
	f, err := os.OpenFile("../../krawler.default.yaml", os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		panic(err)
	}

	err = krawler.DefaultRawConfig.DumpYAML(f)
	if err != nil {
		panic(err)
	}
}
