package job

import (
	"context"
	"fmt"
	"time"
)

type JobI interface {
	Run() error
	Name() string
}

type Job struct {
	JobI
	Time time.Duration
}

type Jobs []Job

func (jobs Jobs) RunLoop(ctx context.Context) {
	for _, job := range jobs {
		fmt.Printf("[JOB] start Job %s\n", job.Name())
		go func(job Job) {
			ticker := time.NewTicker(job.Time)
			defer ticker.Stop()

			if err := job.Run(); err != nil {
				fmt.Printf("[JOB] run job %s error %s\n", job.Name(), err)
			}

			for {
				select {
				case <-ctx.Done():
					return
				case <-ticker.C:
					if err := job.Run(); err != nil {
						fmt.Printf("[JOB] run job %s error %s\n", job.Name(), err)
						continue
					}
				}
			}
		}(job)
	}
}

func (jobs Jobs) RunWithName(name string) error {
	for _, job := range jobs {
		if job.Name() == name {
			return job.Run()
		}
	}
	return nil
}
