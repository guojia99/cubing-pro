package job

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/guojia99/cubing-pro/src/configs"
	"github.com/guojia99/cubing-pro/src/robot/qq_bot/Better-Bot-Go/log"
)

// StaticSiteAutoUpdateJob 与 gateway.StaticSiteConfig 配套：在 RepoDir（或 Root）git pull，有更新则构建。
type StaticSiteAutoUpdateJob struct {
	Site configs.StaticSiteConfig
}

func (j *StaticSiteAutoUpdateJob) Name() string {
	return "StaticSiteAutoUpdate:" + j.Site.StableID()
}

func staticSiteGitDir(site configs.StaticSiteConfig) string {
	r := strings.TrimSpace(site.RepoDir)
	if r != "" {
		return r
	}
	return strings.TrimSpace(site.Root)
}

func gitRevHead(dir string) (string, error) {
	cmd := exec.Command("git", "rev-parse", "HEAD")
	cmd.Dir = dir
	out, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
}

func (j *StaticSiteAutoUpdateJob) Run() error {
	dir := staticSiteGitDir(j.Site)
	id := j.Site.StableID()
	if dir == "" {
		return fmt.Errorf("StaticSiteAutoUpdate[%s]: repo/root dir is empty", id)
	}
	info, err := os.Stat(dir)
	if err != nil || !info.IsDir() {
		return fmt.Errorf("StaticSiteAutoUpdate[%s]: dir not accessible: %s", id, dir)
	}

	before, err := gitRevHead(dir)
	if err != nil {
		return fmt.Errorf("StaticSiteAutoUpdate[%s]: git rev-parse HEAD: %w", id, err)
	}

	pull := exec.Command("git", "pull")
	pull.Dir = dir
	pullOut, err := pull.CombinedOutput()
	if err != nil {
		log.Errorf("StaticSiteAutoUpdate[%s]: git pull failed: %v\n%s", id, err, string(pullOut))
		return fmt.Errorf("StaticSiteAutoUpdate[%s]: git pull: %w", id, err)
	}
	log.Infof("StaticSiteAutoUpdate[%s]: git pull:\n%s", id, strings.TrimSpace(string(pullOut)))

	after, err := gitRevHead(dir)
	if err != nil {
		return fmt.Errorf("StaticSiteAutoUpdate[%s]: git rev-parse after pull: %w", id, err)
	}
	if before == after {
		log.Infof("StaticSiteAutoUpdate[%s]: no new commits, skip build", id)
		return nil
	}

	build := strings.TrimSpace(j.Site.BuildCmd)
	if build == "" {
		build = "npm run build"
	}
	sh := exec.Command("/bin/sh", "-c", build)
	sh.Dir = dir
	sh.Env = os.Environ()
	out, err := sh.CombinedOutput()
	if err != nil {
		log.Errorf("StaticSiteAutoUpdate[%s]: build failed: %v\n%s", id, err, string(out))
		return fmt.Errorf("StaticSiteAutoUpdate[%s]: build: %w", id, err)
	}
	log.Infof("StaticSiteAutoUpdate[%s]: build ok, HEAD %s -> %s\n%s", id, shortHash(before), shortHash(after), strings.TrimSpace(string(out)))
	return nil
}

func shortHash(h string) string {
	h = strings.TrimSpace(h)
	if len(h) <= 7 {
		return h
	}
	return h[:7]
}
