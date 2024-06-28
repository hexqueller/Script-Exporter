package main

import (
	"flag"
	"fmt"
	"github.com/hexqueller/Script-Exporter/internal/config"
	"github.com/hexqueller/Script-Exporter/internal/cron"
	"github.com/hexqueller/Script-Exporter/internal/metrics"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"log"
	"net/http"
)

func main() {
	// получаем аргументы
	port := flag.Int("p", 9105, "Port to listen on")
	configPath := flag.String("c", "configs/default.yaml", "Path to config file")
	debug := flag.Bool("d", false, "Enable debug mode")
	flag.Parse()

	fmt.Println(fmt.Sprintf("Listening on http://localhost:%d/metrics", *port))
	fmt.Println(fmt.Sprintf("Config: %s, Debug is %t", *configPath, *debug))
	fmt.Println()

	// читаем и разбираем файл YAML
	cfg, err := config.LoadConfig(*configPath)
	if err != nil {
		log.Fatalf("error: %v", err)
	}

	// регистрируем метрику
	metrics.RegisterMetrics()

	// создаем и запускаем планировщик
	cron.StartScheduler(debug, cfg.Jobs)

	// создаем HTTP-сервер
	http.Handle("/metrics", promhttp.Handler())
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", *port), nil))
}
