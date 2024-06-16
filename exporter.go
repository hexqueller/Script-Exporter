package main

import (
	"flag"
	"fmt"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"log"
)

type Config struct {
	Jobs map[string]struct {
		Name   string `yaml:"name"`
		Cron   string `yaml:"cron"`
		Script string `yaml:"script"`
	} `yaml:"jobs"`
}

func main() {
	port := flag.Int("p", 9105, "Port to listen on")
	configPath := flag.String("c", "default.yaml", "Path to config file")

	flag.Parse()

	fmt.Println(fmt.Sprintf("Listening on http://localhost:%d", *port))
	fmt.Println(fmt.Sprintf("Config: %s", *configPath))
	fmt.Println("")
	// читаем файл YAML
	data, err := ioutil.ReadFile(*configPath)
	if err != nil {
		log.Fatalf("error: %v", err)
	}

	// разбираем файл YAML в структуру Config
	var config Config
	err = yaml.Unmarshal(data, &config)
	if err != nil {
		log.Fatalf("error: %v", err)
	}

	// выводим все полученные задания
	for _, job := range config.Jobs {
		fmt.Printf("Job: %s\n", job.Name)
		fmt.Printf("Cron: %s\n", job.Cron)
		fmt.Printf("Script: %s\n", job.Script)
		fmt.Println()
	}
}
