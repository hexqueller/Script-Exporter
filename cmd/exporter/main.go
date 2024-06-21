package main

import (
	"Script-Exporter/internal/config"
	"Script-Exporter/internal/cron"
	"Script-Exporter/internal/metrics"
	"flag"
	"fmt"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"log"
	"net/http"
)

func main() {
	// получаем аргументы
	port := flag.Int("p", 9105, "Port to listen on")
	configPath := flag.String("c", "configs/default.yaml", "Path to config file")
	flag.Parse()

	fmt.Println(fmt.Sprintf("Listening on http://localhost:%d/metrics", *port))
	fmt.Println(fmt.Sprintf("Config: %s", *configPath))
	fmt.Println()

	// читаем и разбираем файл YAML
	cfg, err := config.LoadConfig(*configPath)
	if err != nil {
		log.Fatalf("error: %v", err)
	}

	// регистрируем метрику
	metrics.RegisterMetrics()

	// создаем и запускаем планировщик
	cron.StartScheduler(cfg.Jobs)

	// создаем HTTP-сервер
	http.Handle("/metrics", promhttp.Handler())
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", *port), nil))
}
