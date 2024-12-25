### 使用的第三方软件或插件

- [wca-tnoodle](https://github.com/thewca/tnoodle) 用于后端打乱和打乱图片

### 部署

- mysql

```bash
mkdir -p /home/guojia/data/mysql8.0.33

docker run --name mysql8.0.33 \
  -e MYSQL_ALLOW_EMPTY_PASSWORD=yes \
  -p 33306:3306 \
  -v /home/guojia/data/mysql8.0.33:/var/lib/mysql \
  -d mysql:8.0.33

```