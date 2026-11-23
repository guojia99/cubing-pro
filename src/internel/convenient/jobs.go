package convenient

import (
	"slices"
	"strings"
	"time"

	"github.com/guojia99/cubing-pro/src/configs"
	"github.com/guojia99/cubing-pro/src/internel/convenient/job"
	"github.com/guojia99/cubing-pro/src/robot/qq_bot/Better-Bot-Go/log"
	"gorm.io/gorm"
)

func newJobs(db *gorm.DB, runJob bool, config configs.Config) []job.Job {
	var baseJobs = []job.Job{
		{JobI: &job.RecordUpdateJob{DB: db}, Time: time.Minute * 30},
		{JobI: &job.UpdateCubingChinaComps{DB: db}, Time: time.Hour * 24},
	}

	var runJobs []job.Job
	for _, jb := range baseJobs {
		if slices.Contains(config.Job.UnRunJobs, jb.Name()) {
			continue
		}
		runJobs = append(runJobs, jb)
		log.Infof("will run job %s", jb.Name())
	}

	if runJob {
		for _, site := range config.Gateway.StaticSites {
			if !site.AutoUpdate {
				continue
			}
			if strings.TrimSpace(site.Root) == "" {
				log.Errorf("StaticSiteAutoUpdate: skip site %q, root is empty", site.StableID())
				continue
			}
			interval := strings.TrimSpace(site.AutoUpdateInterval)
			d := 5 * time.Minute
			if interval != "" {
				if parsed, err := time.ParseDuration(interval); err != nil {
					log.Errorf("StaticSiteAutoUpdate[%s]: invalid autoUpdateInterval %q, using default 5m: %v", site.StableID(), interval, err)
				} else if parsed > 0 {
					d = parsed
				}
			}
			jobName := (&job.StaticSiteAutoUpdateJob{Site: site}).Name()
			if slices.Contains(config.Job.UnRunJobs, jobName) {
				continue
			}
			runJobs = append(runJobs, job.Job{
				JobI: &job.StaticSiteAutoUpdateJob{Site: site},
				Time: d,
			})
			log.Infof("will run job %s every %s", jobName, d)
		}
	}

	return runJobs
}
