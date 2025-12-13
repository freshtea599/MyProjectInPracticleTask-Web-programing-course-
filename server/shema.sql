-- ================================================================
-- АЙТИ КОФЕЙНЯ - ПОЛНАЯ СХЕМА БД
-- PostgreSQL 18+
-- ================================================================

-- ================================================================
-- РАСШИРЕНИЯ
-- ================================================================

CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- ================================================================
-- ТАБЛИЦА ПОЛЬЗОВАТЕЛЕЙ
-- ================================================================

CREATE TABLE IF NOT EXISTS users (
    id SERIAL PRIMARY KEY,
    username VARCHAR(50) UNIQUE NOT NULL,
    password VARCHAR(255) NOT NULL,
    first_name VARCHAR(100),
    last_name VARCHAR(100),
    birthdate DATE,
    gender CHAR(1),
    profile_tag VARCHAR(50) UNIQUE,
    role VARCHAR(20) NOT NULL DEFAULT 'user',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_users_username ON users(username);
CREATE INDEX idx_users_role ON users(role);
CREATE INDEX idx_users_created_at ON users(created_at DESC);

COMMENT ON TABLE users IS 'Таблица пользователей приложения';
COMMENT ON COLUMN users.role IS 'Роль пользователя: user или admin';
COMMENT ON COLUMN users.gender IS 'Пол: M (мужской), F (женский), O (другое)';

-- ================================================================
-- ТАБЛИЦА ГРУПП ЗАДАЧ
-- ================================================================

CREATE TABLE IF NOT EXISTS groups (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    title VARCHAR(255) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_groups_user_id ON groups(user_id);
CREATE INDEX idx_groups_created_at ON groups(created_at DESC);

COMMENT ON TABLE groups IS 'Группы задач (категории для организации заметок)';
COMMENT ON COLUMN groups.user_id IS 'ID владельца группы';

-- ================================================================
-- ТАБЛИЦА ЗАДАЧ
-- ================================================================

CREATE TABLE IF NOT EXISTS tasks (
    id SERIAL PRIMARY KEY,
    group_id INTEGER NOT NULL REFERENCES groups(id) ON DELETE CASCADE,
    title VARCHAR(255) NOT NULL,
    done BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_tasks_group_id ON tasks(group_id);
CREATE INDEX idx_tasks_done ON tasks(done);
CREATE INDEX idx_tasks_created_at ON tasks(created_at DESC);

COMMENT ON TABLE tasks IS 'Задачи в группах';
COMMENT ON COLUMN tasks.done IS 'Статус выполнения задачи';

-- ================================================================
-- ТАБЛИЦА ТОВАРОВ
-- ================================================================

CREATE TABLE IF NOT EXISTS products (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    price INTEGER NOT NULL,
    image_url VARCHAR(500),
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_products_active ON products(is_active);
CREATE INDEX idx_products_price ON products(price);
CREATE INDEX idx_products_created_at ON products(created_at DESC);

COMMENT ON TABLE products IS 'Каталог товаров (кофейные напитки)';
COMMENT ON COLUMN products.is_active IS 'Товар доступен для покупки';
COMMENT ON COLUMN products.price IS 'Цена в рублях';

-- ================================================================
-- ТАБЛИЦА КОРЗИНЫ
-- ================================================================

CREATE TABLE IF NOT EXISTS cart_items (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    product_id INTEGER NOT NULL,
    name VARCHAR(255) NOT NULL,
    image VARCHAR(500),
    price INTEGER NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_cart_items_user_id ON cart_items(user_id);
CREATE INDEX idx_cart_items_created_at ON cart_items(created_at DESC);

COMMENT ON TABLE cart_items IS 'Товары в корзинах пользователей';
COMMENT ON COLUMN cart_items.product_id IS 'ID товара (может быть удален, но данные сохраняются)';
COMMENT ON COLUMN cart_items.price IS 'Цена товара на момент добавления в корзину';

-- ================================================================
-- ТАБЛИЦА ОТЗЫВОВ
-- ================================================================

CREATE TABLE IF NOT EXISTS reviews (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    rating INTEGER NOT NULL CHECK (rating >= 1 AND rating <= 5),
    comment TEXT NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'pending',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    moderated_by INTEGER REFERENCES users(id) ON DELETE SET NULL,
    moderated_at TIMESTAMP
);

CREATE INDEX idx_reviews_status ON reviews(status);
CREATE INDEX idx_reviews_user_id ON reviews(user_id);
CREATE INDEX idx_reviews_created_at ON reviews(created_at DESC);
CREATE INDEX idx_reviews_moderated_by ON reviews(moderated_by);

COMMENT ON TABLE reviews IS 'Отзывы пользователей о кофейне';
COMMENT ON COLUMN reviews.status IS 'Статус отзыва: pending (ожидает модерации), approved (одобрен), rejected (отклонен)';
COMMENT ON COLUMN reviews.rating IS 'Оценка от 1 до 5 звезд';
COMMENT ON COLUMN reviews.moderated_by IS 'ID администратора, который модерировал отзыв';

-- ================================================================
-- ТРИГГЕРЫ ДЛЯ АВТОМАТИЧЕСКОГО ОБНОВЛЕНИЯ updated_at
-- ================================================================

CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Триггер для users
DROP TRIGGER IF EXISTS update_users_updated_at ON users;
CREATE TRIGGER update_users_updated_at 
BEFORE UPDATE ON users
FOR EACH ROW 
EXECUTE FUNCTION update_updated_at_column();

-- Триггер для groups
DROP TRIGGER IF EXISTS update_groups_updated_at ON groups;
CREATE TRIGGER update_groups_updated_at 
BEFORE UPDATE ON groups
FOR EACH ROW 
EXECUTE FUNCTION update_updated_at_column();

-- Триггер для tasks
DROP TRIGGER IF EXISTS update_tasks_updated_at ON tasks;
CREATE TRIGGER update_tasks_updated_at 
BEFORE UPDATE ON tasks
FOR EACH ROW 
EXECUTE FUNCTION update_updated_at_column();

-- Триггер для products
DROP TRIGGER IF EXISTS update_products_updated_at ON products;
CREATE TRIGGER update_products_updated_at 
BEFORE UPDATE ON products
FOR EACH ROW 
EXECUTE FUNCTION update_updated_at_column();

-- ================================================================
-- ПРЕДСТАВЛЕНИЯ (VIEWS) - опционально для упрощения запросов
-- ================================================================

-- Представление для получения отзывов с данными пользователя
CREATE OR REPLACE VIEW reviews_with_users AS
SELECT 
    r.id,
    r.user_id,
    u.username,
    r.rating,
    r.comment,
    r.status,
    r.created_at,
    r.moderated_by,
    r.moderated_at,
    CASE WHEN r.moderated_by IS NOT NULL 
        THEN (SELECT username FROM users WHERE id = r.moderated_by) 
        ELSE NULL 
    END as moderated_by_username
FROM reviews r
JOIN users u ON r.user_id = u.id;

COMMENT ON VIEW reviews_with_users IS 'Представление отзывов с данными авторов и модераторов';


-- Функция для подсчета отзывов по статусу
CREATE OR REPLACE FUNCTION count_reviews_by_status(status_filter VARCHAR)
RETURNS INTEGER AS $$
DECLARE
    count INTEGER;
BEGIN
    SELECT COUNT(*) INTO count
    FROM reviews
    WHERE status = status_filter;
    RETURN count;
END;
$$ LANGUAGE plpgsql;

-- Функция для получения средней оценки
CREATE OR REPLACE FUNCTION get_average_rating()
RETURNS NUMERIC AS $$
DECLARE
    avg_rating NUMERIC;
BEGIN
    SELECT AVG(rating) INTO avg_rating
    FROM reviews
    WHERE status = 'approved';
    RETURN COALESCE(avg_rating, 0);
END;
$$ LANGUAGE plpgsql;

-- Функция для подсчета активных товаров
CREATE OR REPLACE FUNCTION count_active_products()
RETURNS INTEGER AS $$
DECLARE
    count INTEGER;
BEGIN
    SELECT COUNT(*) INTO count
    FROM products
    WHERE is_active = true;
    RETURN count;
END;
$$ LANGUAGE plpgsql;

