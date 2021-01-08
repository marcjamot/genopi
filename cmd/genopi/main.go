package main

import (
	"errors"
	"flag"
	"log"
	"os"

	"github.com/marcjamot/genopi/internal/common"
	"github.com/marcjamot/genopi/internal/generator"
	"github.com/marcjamot/genopi/internal/parser"
)

func main() {
	log.Print("Genopi - Generate Open API 3")

	status, err := fromFlags()
	if err != nil {
		log.Fatal(err)
	}

	wd, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("- Path: %s", wd)

	log.Print("[1/2] Parse API")
	endpoints, structs, err := parser.FromPath(wd)
	if err != nil {
		log.Fatal(err)
	}

	log.Print("[2/2] Generate documentation")
	if err := generator.OpenAPI3(common.Api{
		Status:    status,
		Endpoints: endpoints,
		Structs:   structs,
	}); err != nil {
		log.Fatal(err)
	}
}

func fromFlags() (common.Status, error) {
	title := flag.String("t", "", "api title")
	version := flag.String("v", "", "api version")
	url := flag.String("u", "", "api url")
	output := flag.String("o", "api.yaml", "output file")

	flag.Parse()

	if *title == "" || *version == "" || *url == "" || *output == "" {
		flag.PrintDefaults()
		return common.Status{}, errors.New("missing required flags")
	}
	return common.Status{
		Title:   *title,
		Version: *version,
		URL:     *url,
		Output:  *output,
	}, nil
}
