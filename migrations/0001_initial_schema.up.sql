-- +migrate Up
CREATE TABLE IF NOT EXISTS users (
    id SERIAL PRIMARY KEY,
    username TEXT UNIQUE NOT NULL,
    email TEXT UNIQUE NOT NULL,
    password TEXT NOT NULL,
    name TEXT NOT NULL,
    photo_url TEXT,
    role TEXT NOT NULL DEFAULT 'user',
    account_status TEXT NOT NULL DEFAULT 'free',
    payment_date TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS locations (
    id SERIAL PRIMARY KEY,
    name TEXT UNIQUE NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS cctvs (
    id SERIAL PRIMARY KEY,
    location_id INTEGER NOT NULL REFERENCES locations(id) ON DELETE CASCADE,
    name TEXT NOT NULL,
    thumbnail_url TEXT,
    source_url TEXT NOT NULL,
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_cctvs_location_id ON cctvs(location_id);
CREATE INDEX IF NOT EXISTS idx_cctvs_is_active ON cctvs(is_active);

INSERT INTO locations (name) VALUES 
('Jakarta'), ('Bogor'), ('Depok'), ('Tangerang'), ('Bekasi'),
('Bandung'), ('Semarang'), ('Yogyakarta'), ('Surabaya'), ('Medan'),
('Makassar'), ('Palembang'), ('Balikpapan'), ('Batam'), ('Pekanbaru')
ON CONFLICT (name) DO NOTHING;

INSERT INTO users (username, email, password, name, role, account_status) VALUES
('admin', 'admin@cctv.app', '$2a$10$X8z5sZ5JZ5Z5Z5Z5Z5Z5.e5Z5Z5Z5Z5Z5Z5Z5Z5Z5Z5Z5Z5Z5', 'Admin', 'admin', 'paid')
ON CONFLICT (username) DO NOTHING;

-- +migrate Down
DROP INDEX IF EXISTS idx_cctvs_location_id;
DROP INDEX IF EXISTS idx_cctvs_is_active;
DROP TABLE IF EXISTS cctvs;
DROP TABLE IF EXISTS locations;
DROP TABLE IF EXISTS users;