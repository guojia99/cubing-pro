# 数据维护脚本

以下示例均从**仓库根目录**执行。脚本用于特定数据维护任务，不是首次启动的必做清单；运行前阅读输入、输出与副作用。应用启动和验证边界见[开发指南](../docs/development.md)。

## MySQL 条件迁移

业务服务初始化会在 `src/internel/convenient/migrate.go` 尝试 AutoMigrate，包括 User 和 OAuthState。下面的 SQL 用于现有 MySQL 库的特定修复或补迁移，**不是所有用户必须执行**，也不适用于直接导入 SQLite。

先备份并核对实际表结构、索引和数据，在隔离库验证后只选择所需脚本：

| 文件 | 适用条件与影响 |
| --- | --- |
| `fix_automigrate_errors.sql` | 旧库 `users.login_id` 空串与唯一约束冲突时，将空串更新为 NULL；`user_kvs.key` 索引修复只写在注释里，直接执行脚本不会执行这些 ALTER |
| `add_wca_auth_columns.sql` | 现有 `users` 表缺少 WCA 登录时间、访问令牌和过期时间字段时补列；没有 `IF NOT EXISTS`，重复运行会报错，部分列已存在时需按实际情况选语句 |
| `add_oauth_states_table.sql` | 缺少 `oauth_states` 时建表；`CREATE TABLE IF NOT EXISTS` 不会修复已存在但结构不匹配的表 |

例如，仅在确认需要补建 OAuth state 表时运行（替换用户名和数据库名，`-p` 交互输入密码）：

```bash
mysql -u YOUR_USER -p YOUR_DATABASE < scripts/add_oauth_states_table.sql
```

其余文件的完整路径为 `scripts/add_wca_auth_columns.sql`、`scripts/fix_automigrate_errors.sql`。不要为消除任意 AutoMigrate 报错而依次盲跑全部 SQL；当前代码忽略迁移错误，启动继续并不能证明表结构已就绪。

## Python 环境

除 `wca_sync_to_sqlite.py` 需要第三方 `requests` 外，当前这些 `.py` 脚本只使用 Python 标准库及同目录模块，无需额外安装 BeautifulSoup、数据库驱动或 pytest。统一使用 Python **3.10 起**可覆盖其中的类型注解与语法；省份统计还依赖 `zoneinfo` 和可用的 IANA 时区数据（默认 `Asia/Shanghai`）。

| 脚本 | 输入/依赖 | 行为与输出 |
| --- | --- | --- |
| `import_local_algs.py` | `test/2x2db/`、`test/algs/` 的 catalog 和原始 JSON 快照 | 导入不含图片载荷的公式 JSON 到 `build/Algs/`，生成 catalog、哈希与来源元数据；`--check` 比对已有文件而不写入 |
| `sync_alg_trainers.py` | 显式传入的固定版本上游 ZIP | 将所选 Trainer JSON 和 LICENSE 写入 `build/Alg-Trainers/`；对 FTO-LT 重复键做脚本定义的合并；`--check` 仅比对 |
| `scrape_2x2db.py` | 网络访问脚本内的 2x2DB 站点 | 抓取站点 JS 数据并按站点展示规则过滤，默认写入 `test/2x2db/`；保留原始/排除条目与来源元数据，不生成图片；数量变化时停止并要求检查 |
| `test_scrape_2x2db.py` | 同目录抓取模块及 `build/Algs/` 已发布 JSON | `unittest` 检查解析、过滤及已发布数据与原始条目的对应；不运行抓取入口，不以 `test/` 原始目录为输入 |
| `db_cubing_wca_id_diff.py` | 默认 `scripts/db.json` 与远程 cubing.com 赛事列表 | 比较赛事 ID，默认打印差异；`--json-out` 额外创建/写入指定 JSON 文件，不连接数据库 |
| `wca_province_stats.py` | cubing.com API、IANA 时区数据 | 分页获取全部 WCA 赛事，在本地按开始日年份统计省份场数/参赛人数；文本或 JSON 输出到标准输出 |
| `wca_city_stats.py` | cubing.com API、`scripts/china_province.data.txt` | 统计各省地级市办赛情况；文本或 JSON 输出到标准输出，无法匹配的数据报告到标准错误 |
| `wca_sync_to_sqlite.py` | Python `requests`、WCA TSV 导出网络访问 | 独立维护离线 SQLite 数据；写入当前工作目录下的 `wca_dbs/`，清理临时文件并删除超出保留数量的旧库，详见下节 |

`db.json` 和 `china_province.data.txt` 是脚本输入数据；它们不会因运行应用自动更新。联网脚本依赖外部响应，文档不保证站点持续可用或字段不变。

## WCA 离线 SQLite 维护

[`wca_sync_to_sqlite.py`](wca_sync_to_sqlite.py) 从 `build/` 原样移入，仅用于独立的离线数据维护，不是应用启动步骤，也不替代服务器 WCA MySQL 查询。服务器查询连接仍由 `src/wca/` 按配置选择 MySQL 库。

该脚本额外依赖 Python 第三方包 `requests`。在选定的 Python 环境中安装后，按需从仓库根目录执行：

```bash
python3 -m pip install requests
python3 scripts/wca_sync_to_sqlite.py
```

- 下载 WCA TSV 导出到固定路径 `/tmp/WCA_export.tsv.zip`；解压前会递归删除已有 `/tmp/wca_export/`，然后将归档解压到该目录。解压成功后删除下载包。
- 输出相对于**进程当前工作目录（cwd）**，不是脚本目录：生成 `./wca_dbs/wca_YYYYMMDD.db` 并写入 `./wca_dbs/sync_status.json`。从仓库根目录运行时输出到仓库根目录的 `wca_dbs/`；迁移不改变这一语义。
- 同步完成后递归删除 `/tmp/wca_export/`；创建数据库失败分支也会删除该解压目录。下载、解压或后续其他步骤失败时不保证清理全部临时文件。
- 最后扫描输出目录中匹配 `wca_*.db` 的文件，按 `os.path.getctime` 降序只保留最新 **2** 个，其余直接删除；这不是按文件名中的日期排序。运行前核对 cwd 和输出目录，需长期保留的数据库应另行备份。
- 下载检查读取 `/tmp/wca_export/metadata.json`，缺失或导出时间超过 26 小时才下载；成功流程会删除该目录，因此不能依赖该检查在每次运行间保留缓存。同一天重复导入可能向已有同名库追加数据。

## 公式导入与来源核对

`test/` 被 `.gitignore` 忽略。`import_local_algs.py` 和它的 `--check` 都先要求 `test/2x2db/catalog.json`、`test/algs/catalog.json` 及对应数据，**公开克隆的导入/来源比对不是自包含流程**。本地恰好存在这些目录不代表它们随仓库发布；`scrape_2x2db.py` 也只提供 2x2DB 抓取，不能代替另一套 `test/algs/` 快照。

只有已准备完整且来源明确的快照时，才执行：

```bash
python3 scripts/import_local_algs.py --check
# 确认需要重建发布数据后执行，会写入 build/Algs/
python3 scripts/import_local_algs.py
```

`sync_alg_trainers.py` 不负责下载 ZIP。当前固定版本为 `3d19a613932319e33219105ad5bfd1f7a128ff24`，要求压缩包内部前缀为 `Alg-Trainers-3d19a613932319e33219105ad5bfd1f7a128ff24/`。准备匹配的上游归档后，把下面的路径替换为实际文件：

```bash
python3 scripts/sync_alg_trainers.py /path/to/Alg-Trainers-3d19a613932319e33219105ad5bfd1f7a128ff24.zip --check
# 确认需要导入后执行，会写入 build/Alg-Trainers/
python3 scripts/sync_alg_trainers.py /path/to/Alg-Trainers-3d19a613932319e33219105ad5bfd1f7a128ff24.zip
```

来源更新应保留许可、版本、哈希与过滤规则，并审阅生成差异。运行时 `algs.Init` 从 `build/Algs/` 及配置的 Trainer 路径加载数据，不依赖原始 `test/` 快照。根忽略规则不再排除整个 `build/Alg-Trainers/`，提交前仍需核对上游文件和实际差异。

## 按需执行示例

以下命令有各自前提，不要当作一段批量执行的初始化脚本。

```bash
# 联网并写入原始快照，默认目录是 test/2x2db/
python3 scripts/scrape_2x2db.py
# 使用仓库中已发布数据执行 Python 检查，不调用抓取入口
python3 -m unittest discover -s scripts -p 'test_scrape_2x2db.py'
# 联网统计，只向标准输出打印 JSON
python3 scripts/wca_province_stats.py --year 2025 --json
python3 scripts/wca_city_stats.py --json
# 联网对比，并创建/写入 local/ 下的结果
python3 scripts/db_cubing_wca_id_diff.py --db scripts/db.json --json-out local/wca-id-diff.json
```

不要把输出重定向到现有源文件；需要保留统计结果时选择明确的本地输出路径。执行结果、网络访问和生成文件应在 PR 的验证说明中如实记录。
