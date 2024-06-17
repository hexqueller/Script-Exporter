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

type Output struct {
	Name     string
	Key      string
	KeyValue string
	Value    string
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
				updateMetrics(parseOutput(string(output)))
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

func parseOutput(output string) map[string]Output {
	result := make(map[string]Output)
	lines := strings.Split(output, "\n")
	for _, line := range lines {
		var (
			name     string
			key      string
			keyValue string
			value    string
		)

		// Split the line into two parts: the metric name and the value
		parts := strings.SplitN(line, " ", 2)
		if len(parts) < 2 {
			log.Printf("Invalid output format: %s", line)
			continue
		}
		metricName := parts[0]
		value = parts[1]

		// Extract the name and key from the metric name
		parts = strings.SplitN(metricName, "{", 2)
		if len(parts) < 2 {
			log.Printf("Invalid output format: %s", line)
			continue
		}
		name = strings.TrimSpace(parts[0])
		keyValuePart := strings.TrimSpace(parts[1])
		keyValuePart = strings.TrimRight(keyValuePart, "}")
		parts = strings.SplitN(keyValuePart, "=", 2)
		if len(parts) < 2 {
			log.Printf("Invalid output format: %s", line)
			continue
		}
		key = strings.TrimSpace(strings.Trim(parts[0], "\""))
		keyValue = strings.TrimSpace(strings.Trim(parts[1], "\""))

		out := Output{Name: name, Key: key, KeyValue: keyValue, Value: value}
		// Use a unique key, such as a combination of name and key
		resultKey := fmt.Sprintf("%s-%s", name, key)
		result[resultKey] = out
	}

	return result
}

func updateMetrics(metrics map[string]Output) {
	for key, out := range metrics {
		fmt.Println(key, out.Name, out.Key, out.KeyValue, out.Value)
		// process the metrics here
	}
}
