package main

import (
	"flag"
	"fmt"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/robfig/cron/v3"
	"gopkg.in/yaml.v2"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strings"
)

type Config struct {
	Jobs []struct {
		Name   string `yaml:"name"`
		Cron   string `yaml:"cron"`
		Script string `yaml:"script"`
	} `yaml:"jobs"`
}

// метрика для результата выполнения скрипта
var scriptResult = prometheus.NewGaugeVec(
	prometheus.GaugeOpts{
		Name: "script_result",
		Help: "Result of script execution",
	},
	[]string{"job"},
)

func main() {
	// получаем аргументы
	port := flag.Int("p", 9105, "Port to listen on")
	configPath := flag.String("c", "default.yaml", "Path to config file")

	flag.Parse()

	fmt.Println(fmt.Sprintf("Listening on http://localhost:%d/metrics", *port))
	fmt.Println(fmt.Sprintf("Config: %s", *configPath))
	fmt.Println()

	// читаем файл YAML
	data, err := os.ReadFile(*configPath)
	if err != nil {
		log.Fatalf("error: %v", err)
	}

	// разбираем файл YAML в структуру Config
	var config Config
	err = yaml.Unmarshal(data, &config)
	if err != nil {
		log.Fatalf("error: %v", err)
	}

	// регистрируем метрику
	prometheus.MustRegister(scriptResult)

	// создаем планировщик
	cron := cron.New()

	// добавляем задания в планировщик
	for _, job := range config.Jobs {
		_, err := cron.AddFunc(job.Cron, func() {
			// выполняем скрипт
			cmd := exec.Command(job.Script)
			output, err := cmd.CombinedOutput()
			if err != nil {
				log.Printf("error running script: %v", err)
				scriptResult.WithLabelValues(job.Name).Set(1)
			} else {
				scriptResult.WithLabelValues(job.Name).Set(0)
				parseOutput(string(output), job.Name)
			}
		})
		if err != nil {
			log.Fatalf("error adding job to cron: %v", err)
		}
	}

	// запускаем планировщик в горутине
	go cron.Start()

	// создаем HTTP-сервер
	http.Handle("/metrics", promhttp.Handler())
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", *port), nil))
}

func parseOutput(output string, jobName string) {
	lines := strings.Split(output, "\n")
	for _, line := range lines {
		var (
			name  string
			key   string
			value string
		)

		// Разбиваем строку на части
		parts := strings.Split(line, "{")
		if len(parts) < 2 {
			log.Printf("Invalid output format: %s", line)
			continue
		}
		name = strings.TrimSpace(parts[0])

		// Разбиваем часть с ключом на части
		keyParts := strings.Split(parts[1], "=")
		if len(keyParts) < 2 {
			log.Printf("Invalid output format: %s", line)
			continue
		}
		key = strings.TrimSpace(strings.Trim(keyParts[0], "\""))
		value = strings.TrimSpace(keyParts[1][:len(keyParts[1])-1]) // удаляем закрывающую скобку

		// Разбиваем значение на части
		valueParts := strings.Split(line, "}")
		if len(valueParts) < 2 {
			log.Printf("Invalid output format: %s", line)
			continue
		}
		value = strings.TrimSpace(valueParts[1])

		fmt.Println("Name:", name)
		fmt.Println("Key:", key)
		fmt.Println("Value:", value)
		fmt.Println()
	}
}
