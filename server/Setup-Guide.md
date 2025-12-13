# Полное руководство по развёртыванию Todo-List сервиса

## Содержание

1. [Предварительные требования](#предварительные-требования)
2. [Установка и настройка](#установка-и-настройка)
3. [Запуск приложения](#запуск-приложения)
4. [Тестирование API](#тестирование-api)
5. [Решение проблем](#решение-проблем)

## Предварительные требования

### Требуемое ПО

- **Go**: 1.20 или выше ([https://golang.org/dl/](https://golang.org/dl/))
- **PostgreSQL**: 18 или выше ([https://www.postgresql.org/download/](https://www.postgresql.org/download/))
- **Git**: для клонирования репозитория

### Проверка установки

```bash
go version
psql --version
git --version
```

## Установка и настройка

### Шаг 1: Клонирование проекта

```bash
git clone https://github.com/yourusername/todolist.git
cd vue/server
```

### Шаг 2: Инициализация Go модуля

Если ещё не инициализировано:

```bash
go mod init vue
go mod tidy
```

Это загрузит все необходимые зависимости:
- `github.com/golang-jwt/jwt/v4` — JWT токены
- `github.com/labstack/echo/v4` — веб-фреймворк
- `github.com/lib/pq` — драйвер PostgreSQL
- `golang.org/x/crypto/bcrypt` — хеширование паролей
- `github.com/joho/godotenv` — загрузка переменных окружения

### Шаг 3: Создание базы данных

Подключитесь к PostgreSQL:

```bash
psql -U postgres
```

Создайте базу данных:

```sql
CREATE DATABASE todolist;
\c todolist
```

### Шаг 4: Создание таблиц

Скопируйте и выполните содержимое файла `schema.sql`:

```bash
psql -U postgres -d todolist -f schema.sql
```

Или выполните вручную в pgAdmin 4.

### Шаг 5: Создание файла .env

В папке `server` создайте файл `.env`:

```bash
touch .env
```

Заполните его следующими переменными:

```env
# Порт сервера
PORT=8080

# Подключение к БД (замените пароль на свой)
DATABASE_URL=postgres://postgres:YOUR_PASSWORD@localhost:5432/todolist?sslmode=require

# JWT секрет (сгенерируйте случайную длинную строку)
JWT_SECRET=your_super_secret_jwt_key_here_change_this_in_production

# Окружение
ENV=development
```

### Шаг 6: Генерирование JWT секрета

Используйте Go для безопасной генерации:

```bash
go run -c "package main; import (\"crypto/rand\"; \"encoding/hex\"; \"fmt\"; \"os\") func main() { b := make([]byte, 32); rand.Read(b); fmt.Println(hex.EncodeToString(b)) }"
```

Или используйте OpenSSL:

```bash
openssl rand -hex 32
```

Скопируйте результат в `JWT_SECRET` в файле `.env`.

### Шаг 7: Добавление .env в .gitignore

Убедитесь, что `.env` не попадает в Git:

```bash
echo ".env" >> .gitignore
git add .gitignore
git commit -m "Add .env to gitignore"
```

## Запуск приложения

### Основная команда

Из папки `server`:

```bash
go run main.go
```

Вы должны увидеть:

```
⇨ http server started on [::]:8080
```

### Запуск в фоновом режиме (на Linux/Mac)

```bash
go run main.go &
```

Или используйте `nohup`:

```bash
nohup go run main.go > server.log 2>&1 &
```

### Запуск с goreman (для desenvolvimento с несколькими процессами)

Установите goreman:

```bash
go install github.com/mattn/goreman@latest
```

Создайте файл `Procfile`:

```
web: go run main.go
```

Запустите:

```bash
goreman start
```

## Тестирование API

### Проверка здоровья сервера

```bash
curl -X GET http://localhost:8080/health
```

Ожидаемый ответ:

```json
{"status": "ok"}
```

### Регистрация нового пользователя

```bash
curl -X POST http://localhost:8080/api/register \
  -H "Content-Type: application/json" \
  -d '{
    "username": "john_doe",
    "password": "SecurePass123!"
  }'
```

Ожидаемый ответ (201 Created):

```json
{
  "id": 1,
  "username": "john_doe",
  "profile_tag": "john_doe_abc123"
}
```

### Авторизация

```bash
curl -X POST http://localhost:8080/api/login \
  -H "Content-Type: application/json" \
  -d '{
    "username": "john_doe",
    "password": "SecurePass123!"
  }'
```

Ожидаемый ответ (200 OK):

```json
{
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "user_id": 1,
  "username": "john_doe"
}
```

**Сохраните токен** — он потребуется для остальных запросов.

### Получение профиля (с авторизацией)

```bash
curl -X GET http://localhost:8080/api/profile \
  -H "Authorization: Bearer YOUR_TOKEN_HERE"
```

## Решение проблем

### Ошибка: "database is locked"

**Причина**: Несколько процессов пытаются получить доступ к БД.

**Решение**:

```bash
# Завершите все процессы PostgreSQL
pkill -f postgres

# Перезагрузите PostgreSQL

### Ошибка: "connection refused"

**Причина**: PostgreSQL не запущена или неверный адрес/порт.

**Решение**:

```bash
# Проверьте статус PostgreSQL
sudo service postgresql status  # Linux

# Или перезагрузите
sudo service postgresql restart

# Проверьте подключение
psql -h localhost -U postgres -d todolist
```

### Ошибка: "invalid token"

**Причина**: JWT токен истёк или некорректен.

**Решение**: Получите новый токен через `/api/login`.

### Ошибка: "permission denied"

**Причина**: Недостаточно прав доступа для администратора.

**Решение**: Убедитесь, что роль пользователя установлена как `admin` в БД.

### Порт уже в использовании

**Причина**: Другой процесс слушает порт 8080.

**Решение**:

```bash
# Найдите процесс
lsof -i :8080

# Завершите его
kill -9 <PID>

# Или используйте другой порт
PORT=3000 go run main.go
```

## Дополнительные ресурсы

- [Go Documentation](https://golang.org/doc/)
- [Echo Framework](https://echo.labstack.com/)
- [PostgreSQL Documentation](https://www.postgresql.org/docs/)
- [JWT Introduction](https://jwt.io/)

---
