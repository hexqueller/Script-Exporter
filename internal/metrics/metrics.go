package metrics

import (
	"fmt"
	"log"
	"sort"
	"strconv"
	"strings"

	"github.com/prometheus/client_golang/prometheus"
)

// Output представляет собой структуру для хранения информации о метрике.
type Output struct {
	Name  string
	Key   map[string]string
	Value string
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
			fmt.Println(fmt.Sprintf("Metric: %s, Key: %v, Value: %s", out.Name, out.Key, out.Value))
		}
		// Бекапим значение
		outCopy := out
		// Регаем метрику
		metric, ok := registeredMetrics[outCopy.Name]
		if !ok {
			// Метрика не существует, создаем новую
			CreateMetric(outCopy.Name, outCopy.Key, outCopy.Value, jobName)
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
			labels := make([]string, 0, len(outCopy.Key))
			for k := range outCopy.Key {
				labels = append(labels, k)
			}
			sort.Strings(labels)
			labelValues := make([]string, 0, len(outCopy.Key))
			for _, label := range labels {
				labelValues = append(labelValues, outCopy.Key[label])
			}
			metric.WithLabelValues(labelValues...).Set(val)
			activeMetrics[jobName][metricKey] = struct{}{}
		}
	}
}

func CreateMetric(name string, key map[string]string, value string, jobName string) {
	metricsHelp := fmt.Sprintf("Job: %s", jobName)
	labelNames := make([]string, 0, len(key))
	for k := range key {
		labelNames = append(labelNames, k)
	}
	sort.Strings(labelNames)
	metric := prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: name,
		Help: metricsHelp,
	}, labelNames)
	registeredMetrics[name] = metric
	prometheus.MustRegister(metric)
	val, err := strconv.ParseFloat(value, 64)
	if err != nil {
		log.Printf("error parsing value to float64: %v", err)
		return
	}
	labels := make([]string, 0, len(key))
	for _, labelName := range labelNames {
		labels = append(labels, key[labelName])
	}
	metric.WithLabelValues(labels...).Set(val)
	activeMetrics[jobName][fmt.Sprintf("%s-%v", name, key)] = struct{}{}
}

func DeleteMetric(jobName string, metricKey string) {
	parts := strings.Split(metricKey, "-")
	if len(parts) != 2 {
		return
	}
	name, keyValueStr := parts[0], parts[1]
	metric, ok := registeredMetrics[name]
	if ok {
		keyValue := make(map[string]string)
		pairs := strings.Split(keyValueStr, " ")
		for _, pair := range pairs {
			parts := strings.SplitN(pair, "=", 2)
			if len(parts) != 2 {
				continue
			}
			key := strings.TrimSpace(parts[0])
			value := strings.TrimSpace(strings.Trim(parts[1], "\""))
			keyValue[key] = value
		}
		labels := make([]string, 0, len(keyValue))
		for _, kv := range keyValue {
			labels = append(labels, kv)
		}
		metric.DeleteLabelValues(labels...)
		log.Printf("Metric deleted: name=%s, keyValue=%v", name, keyValue)
		delete(activeMetrics[jobName], fmt.Sprintf("%s-%v", name, keyValue))
	} else {
		log.Printf("Attempted to delete non-existent metric: name=%s, keyValue=%s", name, keyValueStr)
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
