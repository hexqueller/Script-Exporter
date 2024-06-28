package metrics

import (
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/prometheus/client_golang/prometheus"
)

// Output представляет собой структуру для хранения информации о метрике.
type Output struct {
	Name     string
	Key      string
	KeyValue string
	Value    string
}

var registeredMetrics = make(map[string]*prometheus.GaugeVec)
var activeMetrics = make(map[string]map[string]struct{})

// метрика для результата выполнения скрипта
var scriptResult = prometheus.NewGaugeVec(
	prometheus.GaugeOpts{
		Name: "script_result",
		Help: "Result of script execution",
	},
	[]string{"job"},
)

func RegisterMetrics() {
	prometheus.MustRegister(scriptResult)
}

func UpdateMetrics(metrics map[string]Output, jobName string, debug *bool) {
	for metricKey, out := range metrics {
		if *debug {
			fmt.Println(fmt.Sprintf("Metric: %s, Key: %s, KeyValue: %s, Value: %s", out.Name, out.Key, out.KeyValue, out.Value))
		}
		// Бекапим значение
		outCopy := out
		// Регаем метрику
		metric, ok := registeredMetrics[outCopy.Name]
		if !ok {
			// Метрика не существует, создаем новую
			CreateMetric(outCopy.Name, outCopy.Key, outCopy.KeyValue, outCopy.Value, jobName)
			// Получаем новую метрику
			metric, _ = registeredMetrics[outCopy.Name]
		}
		if metric != nil {
			// Обновляем значение
			val, err := strconv.ParseFloat(outCopy.Value, 64)
			if err != nil {
				log.Printf("error parsing value to float64: %v", err)
				scriptResult.WithLabelValues(jobName).Set(1)
				return
			}
			metric.WithLabelValues(outCopy.KeyValue).Set(val)
			activeMetrics[jobName][metricKey] = struct{}{}
		}
	}
}

func CreateMetric(name string, key string, keyValue string, value string, jobName string) {
	metricsHelp := fmt.Sprintf("Job: %s", jobName)
	metric := prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: name,
		Help: metricsHelp,
	}, []string{key})
	registeredMetrics[name] = metric
	prometheus.MustRegister(metric)
	val, err := strconv.ParseFloat(value, 64)
	if err != nil {
		log.Printf("error parsing value to float64: %v", err)
		return
	}
	metric.WithLabelValues(keyValue).Set(val)
	activeMetrics[jobName][fmt.Sprintf("%s-%s", name, keyValue)] = struct{}{}
}

func DeleteMetric(metricKey string) {
	parts := strings.Split(metricKey, "-")
	if len(parts) != 2 {
		return
	}
	name, keyValue := parts[0], parts[1]
	metric, ok := registeredMetrics[name]
	if ok {
		metric.DeleteLabelValues(keyValue)
		log.Printf("Metric deleted: name=%s, keyValue=%s", name, keyValue)
	} else {
		log.Printf("Attempted to delete non-existent metric: name=%s, keyValue=%s", name, keyValue)
	}
}

func GetActiveMetrics(jobName string) map[string]struct{} {
	return activeMetrics[jobName]
}

func ResetActiveMetrics(jobName string) {
	activeMetrics[jobName] = make(map[string]struct{})
}

func SetScriptResult(jobName string, result float64) {
	scriptResult.WithLabelValues(jobName).Set(result)
}

func IsActiveMetric(jobName, metricKey string) (struct{}, bool) {
	val, exists := activeMetrics[jobName][metricKey]
	return val, exists
}
