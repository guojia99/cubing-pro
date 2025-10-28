#!/usr/bin/env python3
# wca_sync.py - WCA 数据同步脚本

import os
import json
import shutil
import zipfile
import requests
import sqlite3
from datetime import datetime, timedelta
from pathlib import Path

# ================= 配置 =================
DOWNLOAD_URL = "https://www.worldcubeassociation.org/export/results/WCA_export.tsv.zip"
DOWNLOAD_PATH = "/tmp/WCA_export.tsv.zip"
EXTRACT_DIR = "/tmp/wca_export"
DB_OUTPUT_DIR = "./wca_dbs"  # 你可以改成任意路径
SYNC_STATUS_FILE = os.path.join(DB_OUTPUT_DIR, "sync_status.json")

# 保留最新的 N 个数据库文件
KEEP_LATEST_DBS = 2

# ========================================

def parse_export_date(date_str):
    """解析 metadata.json 中的 export_date 字符串为 datetime 对象"""
    # 示例: "2025-10-24 00:00:21 UTC"
    return datetime.strptime(date_str, "%Y-%m-%d %H:%M:%S %Z")

def should_download(metadata_path):
    """检查是否需要重新下载（超过26小时）"""
    if not os.path.exists(metadata_path):
        print("metadata.json 不存在，需要下载。")
        return True

    with open(metadata_path, 'r', encoding='utf-8') as f:
        metadata = json.load(f)

    export_date_str = metadata.get("export_date")
    if not export_date_str:
        print("metadata.json 中缺少 export_date，需要下载。")
        return True

    export_date = parse_export_date(export_date_str)
    now = datetime.utcnow()
    delta = now - export_date

    if delta > timedelta(hours=26):
        print(f"数据已过期 {delta.total_seconds()/3600:.1f} 小时，需要更新。")
        return True
    else:
        print(f"数据在 {delta.total_seconds()/3600:.1f} 小时内，无需更新。")
        return False

def download_and_extract():
    """下载并解压最新数据"""
    print("正在下载 WCA 数据...")
    try:
        response = requests.get(DOWNLOAD_URL, stream=True, timeout=30)
        response.raise_for_status()
        with open(DOWNLOAD_PATH, 'wb') as f:
            for chunk in response.iter_content(chunk_size=8192):
                f.write(chunk)
        print("下载完成。")

        # 清空并创建解压目录
        if os.path.exists(EXTRACT_DIR):
            shutil.rmtree(EXTRACT_DIR)
        os.makedirs(EXTRACT_DIR, exist_ok=True)

        print("正在解压...")
        with zipfile.ZipFile(DOWNLOAD_PATH, 'r') as zip_ref:
            zip_ref.extractall(EXTRACT_DIR)
        print(f"解压到: {EXTRACT_DIR}")

        # 清理下载包
        os.remove(DOWNLOAD_PATH)

        return EXTRACT_DIR
    except Exception as e:
        print(f"下载或解压失败: {e}")
        return None

def create_sqlite_db(extract_dir):
    """将所有 TSV 导入 SQLite，生成 wca_YYYYMMDD.db"""
    # 获取当前日期作为文件名
    date_str = datetime.now().strftime("%Y%m%d")
    db_filename = f"wca_{date_str}.db"
    db_path = os.path.join(DB_OUTPUT_DIR, db_filename)

    os.makedirs(DB_OUTPUT_DIR, exist_ok=True)

    conn = sqlite3.connect(db_path)
    conn.execute("PRAGMA foreign_keys = ON;")

    tsv_files = [f for f in os.listdir(extract_dir) if f.endswith('.tsv')]
    print(f"发现 {len(tsv_files)} 个 TSV 文件。")

    for tsv_file in tsv_files:
        tsv_path = os.path.join(extract_dir, tsv_file)
        table_name = Path(tsv_file).stem  # 去掉 .tsv 作为表名

        print(f"导入 {tsv_file} -> {table_name}")

        # 读取 TSV（跳过表头）
        with open(tsv_path, 'r', encoding='utf-8') as f:
            header = f.readline()
            if not header.strip():
                print(f"警告: {tsv_file} 为空或无表头，跳过。")
                continue
            columns = [col.strip().lower() for col in header.strip().split('\t')]
            # 简单清理列名（防止有特殊字符）
            columns = [col.replace(' ', '_').replace('(', '').replace(')', '') for col in columns]

            # 构造建表语句（全部用 TEXT，简化处理；SQLite 是动态类型）
            col_defs = ", ".join([f"{col} TEXT" for col in columns])
            create_table_sql = f"CREATE TABLE IF NOT EXISTS {table_name} ({col_defs});"
            conn.execute(create_table_sql)

            # 插入数据
            cursor = conn.cursor()
            for line in f:
                if not line.strip():
                    continue
                values = line.strip().split('\t')
                # 处理 NULL 值（MySQL 导出常用 \N 表示 NULL）
                values = [None if v == r'\N' else v for v in values]
                placeholders = ", ".join(['?' for _ in values])
                insert_sql = f"INSERT INTO {table_name} VALUES ({placeholders})"
                try:
                    cursor.execute(insert_sql, values)
                except Exception as e:
                    print(f"插入失败 [{table_name}]: {e}, values: {values}")

        conn.commit()

    conn.close()
    print(f"数据库创建完成: {db_path}")
    return db_path

def update_sync_status(db_path, metadata_path):
    """更新 sync_status.json"""
    status = {
        "latest_db": os.path.basename(db_path),
        "last_sync": datetime.utcnow().strftime("%Y-%m-%d %H:%M:%S UTC"),
        "source_metadata": None
    }

    if os.path.exists(metadata_path):
        with open(metadata_path, 'r', encoding='utf-8') as f:
            status["source_metadata"] = json.load(f)

    with open(SYNC_STATUS_FILE, 'w', encoding='utf-8') as f:
        json.dump(status, f, indent=2, ensure_ascii=False)

    print(f"同步状态已更新: {SYNC_STATUS_FILE}")

def cleanup_old_dbs():
    """清理旧的 .db 文件，只保留最新的两个"""
    db_files = [f for f in os.listdir(DB_OUTPUT_DIR) if f.startswith("wca_") and f.endswith(".db")]
    db_files.sort(key=lambda x: os.path.getctime(os.path.join(DB_OUTPUT_DIR, x)), reverse=True)

    if len(db_files) <= KEEP_LATEST_DBS:
        print(f"保留 {len(db_files)} 个数据库文件，无需清理。")
        return

    for old_db in db_files[KEEP_LATEST_DBS:]:
        old_path = os.path.join(DB_OUTPUT_DIR, old_db)
        os.remove(old_path)
        print(f"已删除旧数据库: {old_path}")

def main():
    metadata_path = os.path.join(EXTRACT_DIR, "metadata.json")

    # 步骤 1: 检查是否需要下载
    if not should_download(metadata_path):
        print("无需更新。")
        return

    # 步骤 2: 下载并解压
    extract_dir = download_and_extract()
    if not extract_dir:
        print("下载失败，终止。")
        return

    # 步骤 3: 导入到 SQLite
    try:
        db_path = create_sqlite_db(extract_dir)
    except Exception as e:
        print(f"创建数据库失败: {e}")
        shutil.rmtree(extract_dir)
        return

    # 步骤 4: 更新同步状态
    update_sync_status(db_path, os.path.join(extract_dir, "metadata.json"))

    # 步骤 5: 清理临时文件
    shutil.rmtree(extract_dir)
    print("临时文件已删除。")

    # 步骤 6: 清理旧数据库
    cleanup_old_dbs()

    print("✅ 同步完成。")

if __name__ == "__main__":
    main()