-- migrate:up

-- Создание схемы 'shop'
CREATE SCHEMA IF NOT EXISTS shop;

-- Таблица пользователей (User)
CREATE TABLE shop.users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email VARCHAR(255) NOT NULL UNIQUE,
    role VARCHAR(50) CHECK (role IN ('employee', 'moderator')) NOT NULL,
    password_hash CHAR(64) DEFAULT NULL,
    created_at TIMESTAMP DEFAULT NOW()
);

-- Таблица ПВЗ (PVZ)
CREATE TABLE shop.pvz (
    id UUID PRIMARY KEY,
    city VARCHAR(100) CHECK (city IN ('Москва', 'Санкт-Петербург', 'Казань')) NOT NULL,
    registration_date TIMESTAMP NOT NULL
);

-- Таблица приемок (Reception)
CREATE TABLE shop.receptions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    pvz_id UUID REFERENCES shop.pvz(id),
    status VARCHAR(50) CHECK (status IN ('in_progress', 'close')) NOT NULL,
    created_at TIMESTAMP DEFAULT NOW()
);

-- Таблица товаров (Product)
CREATE TABLE shop.products (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    type VARCHAR(50) CHECK (type IN ('электроника', 'одежда', 'обувь')) NOT NULL,
    reception_id UUID REFERENCES shop.receptions(id),
    created_at TIMESTAMP DEFAULT NOW()
);

-- Индексы для ускорения запросов
CREATE INDEX idx_users_email ON shop.users(email);
CREATE INDEX idx_receptions_pvz_id_created_at ON shop.receptions (pvz_id, created_at);


-- migrate:down
-- Индексы
DROP INDEX IF EXISTS idx_users_email;
DROP INDEX IF EXISTS idx_receptions_pvz_id_created_at;

-- Удаление схемы
DROP SCHEMA IF EXISTS shop CASCADE;
