
# Айти кофейня — полная инструкция

## Развертывание сервера и клиента

---

## Оглавление
1. Введение  
2. Требования  
3. Установка сервера (Go + PostgreSQL)  
4. Установка клиента (Vue 3 + Vite)  
5. Запуск приложения  
6. API документация  
7. Развертывание в продакшене  
8. Troubleshooting  

---

## 1. Введение

**Айти кофейня** — полнофункциональное веб-приложение для управления магазином кофейных напитков, пользовательскими заметками и модерацией отзывов.

### Архитектура проекта

- **Backend:** Go + Echo + PostgreSQL  
- **Frontend:** Vue 3 + Vite + Bootstrap 5  

---

## 2. Требования

### Для сервера
- Go 1.21+  
- PostgreSQL 18+  
- Git  
- Текстовый редактор  

### Для клиента
- Node.js 18+  
- npm 8+ или yarn 1.22+  
- Современный браузер  


---

## 3. Установка сервера (Go + PostgreSQL)

### Шаг 1. Клонирование репозитория

```bash
git clone <https://github.com/freshtea599/MyProjectInPracticleTask-Web-programing-course-.git>
cd server
```

### Шаг 2. Установка зависимостей

```bash
go mod download
go mod tidy
```

### Шаг 3. Установка PostgreSQL

**macOS (Homebrew):**
```bash
brew install postgresql
brew services start postgresql
```

**Ubuntu:**
```bash
sudo apt update
sudo apt install postgresql postgresql-contrib
sudo systemctl start postgresql
```

### Шаг 4. Создание базы данных

```bash
psql -U postgres
CREATE DATABASE aitikofeynya;
\q
```

### Шаг 5. Применение схемы БД (лежит в папке сервера)

```bash
psql -U postgres -d aitikofeynya -f schema.sql
```

### Шаг 6. Файл `.env`

```env
DATABASE_URL="postgres://postgres:_Vash_Porol_@localhost:5432/todolist?sslmode=disable"
JWT_SECRET="your-super-secret-jwt-key-change-in-production-min-32-chars"
PORT=8080
```

Генерация ключа:
```bash
openssl rand -base64 32
```

### Шаг 7. Назначение администратора

```sql
UPDATE users SET role = 'admin' WHERE username = 'testuser';
```

### Шаг 8. Запуск сервера

```bash
go run main.go
```

Проверка:
```bash
curl http://localhost:8080/health
```

---

## 4. Установка клиента (Vue 3 + Vite)

### Шаг 1. Переход в папку клиента

```bash
cd ../client
```

### Шаг 2. Установка зависимостей

```bash
npm install
# или
yarn install
```



### Шаг 3. Запуск

```bash
npm run dev
```

---

## 4. Запуск приложения

1. Запустить PostgreSQL  
2. Сервер:
```bash
go run main.go
```
3. Клиент:
```bash
npm run dev
```
4. Открыть: http://localhost:5173  

---

## 6. API документация

### Публичные эндпоинты

**POST /api/register**
```json
{
  "username": "newuser",
  "password": "securePassword123"
}
```

**POST /api/login**
```json
{
  "username": "newuser",
  "password": "securePassword123"
}
```

### Защищенные эндпоинты

Заголовок:
```
Authorization: Bearer <jwt>
```

**GET /api/profile**  
**PUT /api/profile**  

### Админские эндпоинты

- `/api/admin/products`
- `/api/admin/reviews`

---

## 7. Продакшен

Рекомендуется:
- Nginx как reverse proxy  
- HTTPS (Let's Encrypt)  
- Переменные окружения  
- Docker / docker-compose  

---

## 8. Troubleshooting

- Проверяйте `.env` файлы  
- Убедитесь, что порты не заняты  
- Проверяйте логи сервера  

---

