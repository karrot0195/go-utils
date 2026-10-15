package main

import (
	"flag"
	"fmt"

	"go-utils/packages/openroute"
)

func main() {
	modelType := flag.String("type", "free", "type of models: free | all")
	flag.Parse()

	data, cacheFile, err := openroute.FetchModels()
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	fmt.Println("Using cache file:", cacheFile)

	resp, err := openroute.ParseModels(data)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	openroute.PrintModels(resp, *modelType)
}
