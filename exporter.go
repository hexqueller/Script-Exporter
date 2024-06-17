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
			executeScriptAndUpdateMetrics(job.Name, job.Script)
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

func executeScriptAndUpdateMetrics(jobName string, script string) {
	// выполняем скрипт
	cmd := exec.Command(script)
	output, err := cmd.CombinedOutput()
	if err != nil {
		log.Printf("error running script: %v", err)
		scriptResult.WithLabelValues(jobName).Set(1)
	} else {
		scriptResult.WithLabelValues(jobName).Set(0)
		var outputs []string
		outputs = strings.Split(string(output), "\n")
		for _, out := range outputs {
			if len(out) > 0 {
				updateMetrics(parseOutput(out), jobName)
			}
		}
	}
}

func parseOutput(output string) map[string]Output {
	result := make(map[string]Output)
	line := strings.TrimSpace(strings.ReplaceAll(output, "\r", ""))
	if len(line) == 0 {
		return result
	}

	openBrace := strings.Index(line, "{")
	closeBrace := strings.Index(line, "}")
	if openBrace == -1 || closeBrace == -1 {
		log.Printf("Invalid output format: %s", line)
		return result
	}

	name := strings.TrimSpace(line[:openBrace])
	keyValue := strings.TrimSpace(line[openBrace+1 : closeBrace])
	value := strings.TrimSpace(line[closeBrace+1:])
	keyValueParts := strings.SplitN(keyValue, "=", 2)
	if len(keyValueParts) != 2 {
		log.Printf("Invalid output format: %s", line)
		return result
	}
	key := strings.TrimSpace(keyValueParts[0])
	keyValue = strings.TrimSpace(keyValueParts[1])

	out := Output{Name: name, Key: key, KeyValue: keyValue, Value: value}
	var resultKey string
	resultKey = fmt.Sprintf("%s-%s", name, key)
	result[resultKey] = out

	return result
}

func updateMetrics(metrics map[string]Output, jobName string) {
	for _, out := range metrics {
		fmt.Println(out.Name, out.Key, out.KeyValue, out.Value)
		// дальше будет проверка
	}
}
