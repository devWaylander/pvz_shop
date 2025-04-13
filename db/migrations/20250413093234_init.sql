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
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    city VARCHAR(100) CHECK (city IN ('Москва', 'Санкт-Петербург', 'Казань')) NOT NULL,
    created_at TIMESTAMP DEFAULT NOW()
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
CREATE INDEX idx_shop_pvz_created_at ON shop.pvz (created_at);
CREATE INDEX idx_shop_reception_created_at ON shop.receptions (created_at);
CREATE INDEX idx_shop_product_created_at ON shop.products (created_at);


-- migrate:down
-- Индексы
DROP INDEX IF EXISTS shop.idx_shop_pvz_registration_date;
DROP INDEX IF EXISTS shop.idx_shop_reception_date_time;
DROP INDEX IF EXISTS shop.idx_shop_product_date_time;

-- Удаление схемы
DROP SCHEMA IF EXISTS shop CASCADE;
