package cron

import (
	"github.com/hexqueller/Script-Exporter/internal/script"
	"log"

	"github.com/robfig/cron/v3"
)

func StartScheduler(jobs []struct {
	Name   string `yaml:"name"`
	Cron   string `yaml:"cron"`
	Script string `yaml:"script"`
}) {
	cronInstance := cron.New()
	for _, job := range jobs {
		job := job
		_, err := cronInstance.AddFunc(job.Cron, func() {
			script.ExecuteScriptAndUpdateMetrics(job.Name, job.Script)
		})
		if err != nil {
			log.Fatalf("error adding job to cron: %v", err)
		}
	}
	go cronInstance.Start()
}
