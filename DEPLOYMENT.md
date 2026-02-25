# Руководство по развёртыванию

## Требования к серверу

### Минимальные требования
- CPU: 2 ядра
- RAM: 4 GB
- Диск: 20 GB SSD
- ОС: Ubuntu 20.04+ / Debian 11+ / CentOS 8+

### Рекомендуемые требования
- CPU: 4 ядра
- RAM: 8 GB
- Диск: 50 GB SSD
- ОС: Ubuntu 22.04 LTS

## Установка зависимостей

### 1. Установка Go

```bash
wget https://go.dev/dl/go1.23.0.linux-amd64.tar.gz
sudo rm -rf /usr/local/go
sudo tar -C /usr/local -xzf go1.23.0.linux-amd64.tar.gz
echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.bashrc
source ~/.bashrc
go version
```

### 2. Установка PostgreSQL 16

```bash
sudo apt update
sudo apt install -y postgresql-16 postgresql-contrib-16

# Настройка PostgreSQL
sudo -u postgres psql
```

```sql
CREATE DATABASE oculus_db;
CREATE USER oculus_user WITH ENCRYPTED PASSWORD 'secure_password';
GRANT ALL PRIVILEGES ON DATABASE oculus_db TO oculus_user;
\q
```

### 3. Установка Redis

```bash
sudo apt install -y redis-server
sudo systemctl enable redis-server
sudo systemctl start redis-server
```

### 4. Установка MinIO (опционально)

```bash
wget https://dl.min.io/server/minio/release/linux-amd64/minio
chmod +x minio
sudo mv minio /usr/local/bin/

# Создание systemd сервиса
sudo tee /etc/systemd/system/minio.service > /dev/null <<EOF
[Unit]
Description=MinIO
After=network.target

[Service]
Type=simple
User=minio
Group=minio
ExecStart=/usr/local/bin/minio server /data/minio --console-address ":9001"
Restart=always

[Install]
WantedBy=multi-user.target
EOF

sudo useradd -r -s /bin/false minio
sudo mkdir -p /data/minio
sudo chown minio:minio /data/minio
sudo systemctl enable minio
sudo systemctl start minio
```

## Развёртывание приложения

### 1. Клонирование репозитория

```bash
cd /opt
sudo git clone <repository-url> oculus-backend
cd oculus-backend
```

### 2. Настройка окружения

```bash
sudo cp .env.example .env
sudo nano .env
```

Пример production конфигурации:

```env
# Server
PORT=8080
GIN_MODE=release

# Database
DB_HOST=localhost
DB_PORT=5432
DB_USER=oculus_user
DB_PASSWORD=secure_password
DB_NAME=oculus_db
DB_SSLMODE=require

# JWT (сгенерируйте надёжный ключ)
JWT_SECRET=your-very-long-and-secure-secret-key-change-this
JWT_ACCESS_EXPIRY=15m
JWT_REFRESH_EXPIRY=168h

# Redis
REDIS_HOST=localhost
REDIS_PORT=6379
REDIS_PASSWORD=
REDIS_DB=0

# MinIO
MINIO_ENDPOINT=localhost:9000
MINIO_ACCESS_KEY=minioadmin
MINIO_SECRET_KEY=minioadmin
MINIO_BUCKET=oculus-media
MINIO_USE_SSL=false

# Telegram
TELEGRAM_BOT_TOKEN=your-bot-token-from-botfather

# Logging
LOG_LEVEL=info
```

### 3. Сборка приложения

```bash
cd /opt/oculus-backend
go build -o bin/api ./cmd/api
go build -o bin/seed ./cmd/seed
```

### 4. Создание systemd сервиса

```bash
sudo tee /etc/systemd/system/oculus-api.service > /dev/null <<EOF
[Unit]
Description=Oculus API Server
After=network.target postgresql.service redis.service

[Service]
Type=simple
User=oculus
Group=oculus
WorkingDirectory=/opt/oculus-backend
ExecStart=/opt/oculus-backend/bin/api
Restart=always
RestartSec=5
StandardOutput=journal
StandardError=journal

[Install]
WantedBy=multi-user.target
EOF

# Создание пользователя
sudo useradd -r -s /bin/false oculus
sudo chown -R oculus:oculus /opt/oculus-backend

# Запуск сервиса
sudo systemctl daemon-reload
sudo systemctl enable oculus-api
sudo systemctl start oculus-api
```

### 5. Проверка статуса

```bash
sudo systemctl status oculus-api
sudo journalctl -u oculus-api -f
```

## Настройка Nginx

### 1. Установка Nginx

```bash
sudo apt install -y nginx
```

### 2. Конфигурация

```bash
sudo tee /etc/nginx/sites-available/oculus > /dev/null <<EOF
upstream oculus_backend {
    server 127.0.0.1:8080;
}

server {
    listen 80;
    server_name your-domain.com;

    client_max_body_size 20M;

    location / {
        proxy_pass http://oculus_backend;
        proxy_http_version 1.1;
        proxy_set_header Upgrade \$http_upgrade;
        proxy_set_header Connection 'upgrade';
        proxy_set_header Host \$host;
        proxy_set_header X-Real-IP \$remote_addr;
        proxy_set_header X-Forwarded-For \$proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto \$scheme;
        proxy_cache_bypass \$http_upgrade;
    }

    location /api/v1/media/upload {
        proxy_pass http://oculus_backend;
        proxy_request_buffering off;
        client_max_body_size 20M;
    }
}
EOF

sudo ln -s /etc/nginx/sites-available/oculus /etc/nginx/sites-enabled/
sudo nginx -t
sudo systemctl restart nginx
```

### 3. Настройка SSL (Let's Encrypt)

```bash
sudo apt install -y certbot python3-certbot-nginx
sudo certbot --nginx -d your-domain.com
sudo systemctl reload nginx
```

## Мониторинг и логирование

### 1. Просмотр логов

```bash
# Логи приложения
sudo journalctl -u oculus-api -f

# Логи Nginx
sudo tail -f /var/log/nginx/access.log
sudo tail -f /var/log/nginx/error.log

# Логи PostgreSQL
sudo tail -f /var/log/postgresql/postgresql-16-main.log
```

### 2. Мониторинг ресурсов

```bash
# CPU и память
htop

# Дисковое пространство
df -h

# Сетевые соединения
netstat -tulpn | grep :8080
```

## Резервное копирование

### 1. Резервное копирование базы данных

```bash
# Создание backup скрипта
sudo tee /opt/backup-db.sh > /dev/null <<'EOF'
#!/bin/bash
BACKUP_DIR="/backup/postgres"
DATE=$(date +%Y%m%d_%H%M%S)
mkdir -p $BACKUP_DIR

pg_dump -U oculus_user -h localhost oculus_db | gzip > $BACKUP_DIR/oculus_db_$DATE.sql.gz

# Удаление старых backup (старше 7 дней)
find $BACKUP_DIR -name "*.sql.gz" -mtime +7 -delete
EOF

sudo chmod +x /opt/backup-db.sh

# Добавление в cron (ежедневно в 2:00)
echo "0 2 * * * /opt/backup-db.sh" | sudo crontab -
```

### 2. Резервное копирование медиафайлов

```bash
# Backup MinIO данных
sudo tee /opt/backup-media.sh > /dev/null <<'EOF'
#!/bin/bash
BACKUP_DIR="/backup/media"
DATE=$(date +%Y%m%d_%H%M%S)
mkdir -p $BACKUP_DIR

tar -czf $BACKUP_DIR/media_$DATE.tar.gz /data/minio

# Удаление старых backup (старше 30 дней)
find $BACKUP_DIR -name "*.tar.gz" -mtime +30 -delete
EOF

sudo chmod +x /opt/backup-media.sh
echo "0 3 * * 0 /opt/backup-media.sh" | sudo crontab -
```

## Обновление приложения

### 1. Обновление кода

```bash
cd /opt/oculus-backend
sudo -u oculus git pull origin main
sudo -u oculus go build -o bin/api ./cmd/api
sudo systemctl restart oculus-api
```

### 2. Откат к предыдущей версии

```bash
cd /opt/oculus-backend
sudo -u oculus git log --oneline -10
sudo -u oculus git checkout <commit-hash>
sudo -u oculus go build -o bin/api ./cmd/api
sudo systemctl restart oculus-api
```

## Безопасность

### 1. Настройка файрвола

```bash
sudo ufw allow 22/tcp
sudo ufw allow 80/tcp
sudo ufw allow 443/tcp
sudo ufw enable
```

### 2. Ограничение доступа к PostgreSQL

```bash
sudo nano /etc/postgresql/16/main/pg_hba.conf
```

Добавьте:
```
host    oculus_db    oculus_user    127.0.0.1/32    scram-sha-256
```

### 3. Настройка fail2ban

```bash
sudo apt install -y fail2ban

sudo tee /etc/fail2ban/jail.local > /dev/null <<EOF
[DEFAULT]
bantime = 3600
findtime = 600
maxretry = 5

[sshd]
enabled = true
EOF

sudo systemctl enable fail2ban
sudo systemctl start fail2ban
```

## Масштабирование

### Горизонтальное масштабирование

1. **Load Balancer**: Используйте Nginx или HAProxy для распределения нагрузки
2. **Несколько инстансов**: Запустите несколько копий приложения на разных портах
3. **Shared Redis**: Все инстансы должны использовать один Redis
4. **Shared PostgreSQL**: Все инстансы подключаются к одной БД

Пример конфигурации Nginx для балансировки:

```nginx
upstream oculus_backend {
    least_conn;
    server 127.0.0.1:8080;
    server 127.0.0.1:8081;
    server 127.0.0.1:8082;
}
```

### Вертикальное масштабирование

1. **PostgreSQL**: Увеличьте `shared_buffers`, `work_mem`, `max_connections`
2. **Redis**: Настройте `maxmemory` и политику вытеснения
3. **Go**: Установите `GOMAXPROCS` для использования всех ядер

## Troubleshooting

### Приложение не запускается

```bash
# Проверка логов
sudo journalctl -u oculus-api -n 100

# Проверка портов
sudo netstat -tulpn | grep :8080

# Проверка прав доступа
ls -la /opt/oculus-backend
```

### Проблемы с базой данных

```bash
# Проверка подключения
psql -U oculus_user -h localhost -d oculus_db

# Проверка активных соединений
sudo -u postgres psql -c "SELECT * FROM pg_stat_activity;"
```

### Высокая нагрузка

```bash
# Анализ медленных запросов
sudo -u postgres psql oculus_db -c "SELECT * FROM pg_stat_statements ORDER BY total_time DESC LIMIT 10;"

# Мониторинг памяти
free -h
```

## Контакты поддержки

При возникновении проблем обращайтесь к команде разработки.
