package metrics

import (
	"fmt"
	"github.com/prometheus/client_golang/prometheus"
	"log"
	"sort"
	"strconv"
	"strings"
)

// Output представляет собой структуру для хранения информации о метрике.
type Output struct {
	Name   string
	Labels map[string]string
	Value  string
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
			fmt.Println(fmt.Sprintf("Metric: %s, Labels: %v, Value: %s", out.Name, out.Labels, out.Value))
		}
		// Бекапим значение
		outCopy := out
		// Регаем метрику
		metric, ok := registeredMetrics[outCopy.Name]
		if !ok {
			// Метрика не существует, создаем новую
			CreateMetric(outCopy.Name, outCopy.Labels, outCopy.Value, jobName)
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
			labels := make([]string, 0, len(outCopy.Labels))
			for k := range outCopy.Labels {
				labels = append(labels, k)
			}
			sort.Strings(labels)
			labelValues := make([]string, 0, len(outCopy.Labels))
			for _, label := range labels {
				labelValues = append(labelValues, outCopy.Labels[label])
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

func DeleteMetric(debug *bool, metricName string, labels map[string]string) {
	metric, ok := registeredMetrics[metricName]
	if ok {
		if metric.Delete(labels) {
			if *debug {
				log.Printf("Metric deleted: name=%s, keyValue=%s", metricName, labels)
			}
		} else {
			log.Printf("Metric found but labels is invalid: name=%s, keyValue=%s", metricName, labels)
		}
	} else {
		log.Printf("Attempted to delete non-existent metric: name=%s, keyValue=%s", metricName, labels)
	}
}

func ParseMetricToDelete(debug *bool, metricString string) (*bool, string, map[string]string) {
	parts := strings.Split(metricString, "-map[")
	if len(parts) != 2 {
		return debug, "", nil
	}

	metricName := strings.TrimSpace(parts[0])
	labelsStr := strings.TrimSpace(parts[1])
	labelsStr = strings.Trim(labelsStr, "]")
	labelParts := strings.Split(labelsStr, " ")

	labels := make(map[string]string)
	for _, labelPart := range labelParts {
		keyValue := strings.Split(labelPart, ":")
		if len(keyValue) != 2 {
			continue
		}
		labels[keyValue[0]] = keyValue[1]
	}

	return debug, metricName, labels
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
