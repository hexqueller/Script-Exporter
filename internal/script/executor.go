package script

import (
	"fmt"
	"log"
	"os/exec"
	"strings"

	"github.com/hexqueller/Script-Exporter/internal/metrics"
)

func ExecuteScriptAndUpdateMetrics(jobName string, script string) {
	// сохраняем текущий список активных метрик
	oldActiveMetrics := metrics.GetActiveMetrics(jobName)
	metrics.ResetActiveMetrics(jobName)

	// определяем тип скрипта на основе расширения файла
	var cmd *exec.Cmd
	if strings.HasSuffix(script, ".sh") {
		// запуск скрипта bash
		cmd = exec.Command("bash", "-c", script)
	} else if strings.HasSuffix(script, ".py") {
		// запуск скрипта Python
		cmd = exec.Command("python3", script)
	} else {
		log.Printf("Unsupported script type: %s", script)
		metrics.SetScriptResult(jobName, 1)
		return
	}

	output, err := cmd.CombinedOutput()
	if err != nil {
		log.Printf("error running script: %v", err)
		metrics.SetScriptResult(jobName, 1)
	} else {
		metrics.SetScriptResult(jobName, 0)
		var outputs []string
		outputs = strings.Split(string(output), "\n")
		for _, out := range outputs {
			if len(out) > 0 {
				metrics.UpdateMetrics(parseOutput(out, jobName), jobName)
			}
		}
		fmt.Println()
	}

	// удаляем пропавшие метрики
	for metricKey := range oldActiveMetrics {
		if _, exists := metrics.IsActiveMetric(jobName, metricKey); !exists {
			metrics.DeleteMetric(metricKey)
		}
	}
}

func parseOutput(output string, jobName string) map[string]metrics.Output {
	result := make(map[string]metrics.Output)
	line := strings.TrimSpace(strings.ReplaceAll(output, "\r", ""))
	if len(line) == 0 {
		return result
	}

	openBrace := strings.Index(line, "{")
	closeBrace := strings.Index(line, "}")
	if openBrace == -1 || closeBrace == -1 {
		log.Printf("Invalid output format: %s", line)
		metrics.SetScriptResult(jobName, 1)
		return result
	}

	name := strings.TrimSpace(line[:openBrace])
	keyValue := strings.TrimSpace(line[openBrace+1 : closeBrace])
	value := strings.TrimSpace(line[closeBrace+1:])
	keyValueParts := strings.SplitN(keyValue, "=", 2)
	if len(keyValueParts) != 2 {
		log.Printf("Invalid output format: %s", line)
		metrics.SetScriptResult(jobName, 1)
		return result
	}
	key := strings.TrimSpace(keyValueParts[0])
	keyValue = strings.TrimSpace(strings.Trim(keyValueParts[1], "\""))

	out := metrics.Output{Name: name, Key: key, KeyValue: keyValue, Value: value}
	var resultKey string
	resultKey = fmt.Sprintf("%s-%s", name, keyValue)
	result[resultKey] = out

	return result
}
