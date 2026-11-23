# README 展示与维护 / README widgets

[中文 README](../README.md) · [English README](../README.en.md)

## 展示来源 / Sources

| 展示 / Widget | 来源 / Source | 凭据 / Credentials |
| --- | --- | --- |
| Stars、Forks、贡献者数量、最近提交 / Repository badges | Shields.io 的公开仓库接口 / Public repository endpoints | 无 / None |
| 贡献者头像墙 / Contributor wall | `contrib.rocks/image?repo=guojia99/cubing-pro` | 无 / None |
| Star 趋势 / Star history | `api.star-history.com`，支持深浅色 / Supports light and dark themes | 无 / None |
| 仓库活跃曲线 / Repository activity | `docs/assets/activity.svg`，从完整 Git 历史生成 / Generated from full Git history | 无外部 API 密钥 / No external API key |
| 头图 / Banner | `docs/assets/readme-banner.svg`，仓库内置 SVG / Bundled SVG | 无 / None |

远程组件受到第三方服务可用性和限流影响。两份 README 均保留 GitHub 原始页面链接。头像墙仅展示 GitHub 可识别的代码贡献者，不代表全部社区贡献。

Remote widgets depend on third-party availability and rate limits. Both READMEs link to the original GitHub pages as fallbacks. The avatar wall shows code contributors recognized by GitHub, rather than every community contribution.

## 活跃曲线 / Activity chart

脚本仅使用 Python 3 标准库与 Git，从当前 `HEAD` 可达历史中读取非合并提交，以提交者时间按 UTC 周一开始的自然周统计，展示当前周及之前 25 周。当前日期之后的记录不计入。作者按 `%aE`（mailmap 归一后的邮箱，不区分大小写）去重；GitHub Actions 图表机器人的固定邮箱被排除，邮箱不会写入图片。此口径不同于 GitHub 的贡献者与贡献日历。

The script uses only Python 3's standard library and Git. It reads non-merge commits reachable from `HEAD`, groups committer dates into UTC weeks starting on Monday, and shows the current week plus the previous 25. Records after the cutoff date are excluded. Authors are deduplicated by case-insensitive `%aE` (mailmap-normalized email). The chart bot's fixed GitHub Actions email is excluded; email addresses are never written to the image. These metrics differ from GitHub's contributor counts and contribution calendar.

在仓库根目录生成；只改写图表文件，不连接数据库或启动 Go 服务：

Run from the repository root; this only rewrites the chart, without accessing databases or starting Go services:

```bash
python3 scripts/update_readme_activity.py

# 可重复生成指定截止日期的快照 / Reproduce a snapshot for a cutoff date
python3 scripts/update_readme_activity.py --as-of 2026-11-23
git diff --check
```

浅克隆需先运行 `git fetch --unshallow`。曲线的零值代表该周没有符合统计条件的提交，不代表数据丢失。

For a shallow clone, run `git fetch --unshallow` first. A zero means no qualifying commits in that week, rather than missing data.

## 自动更新 / Automatic updates

[工作流](../.github/workflows/readme-activity.yml)在默认分支每周一 **10:17（Asia/Shanghai，UTC+8）**计划运行；GitHub 的定时任务可能延迟。也可在 **Actions → Update README activity → Run workflow** 手动触发。任务拉取默认分支的完整历史，仅在图表有变化时提交 `docs/assets/activity.svg`。机器人提交带 `[skip ci]`，且被图表统计排除。

The [workflow](../.github/workflows/readme-activity.yml) is scheduled on the default branch every Monday at **02:17 UTC** (10:17 Asia/Shanghai). GitHub may delay scheduled runs. Trigger it manually under **Actions → Update README activity → Run workflow**. It checks out the default branch with full history and commits only `docs/assets/activity.svg` when changed. Bot commits include `[skip ci]` and are excluded from the chart.

工作流使用仓库自带 `GITHUB_TOKEN` 的 `contents: write` 权限，不需要配置个人令牌。若仓库策略禁止 Actions 写入，或分支保护拒绝机器人推送，任务会失败；按仓库政策调整工作流权限，或本地运行脚本后通过 PR 提交图表。不要为更新图表关闭分支保护。公开仓库长期无活动时 GitHub 可能停用定时工作流，届时在 Actions 中重新启用。

The workflow uses the repository's built-in `GITHUB_TOKEN` with `contents: write`; no personal token is required. If repository policy denies writes or branch protection rejects the bot's push, the run fails. Follow repository policy when configuring workflow permissions, or generate the chart locally and submit it through a PR. Keep branch protections in place. GitHub may disable scheduled workflows after prolonged inactivity in public repositories; re-enable the workflow in Actions if needed.

迁移仓库时，同步修改两份 README 的仓库链接、远程图片参数和工作流的 `github.repository` 条件。更新文案、功能说明或启动命令时同步维护两种语言。

When moving the repository, update repository links and remote image parameters in both READMEs, along with the workflow's `github.repository` condition. Keep both languages aligned when changing descriptions, features, or setup commands.
