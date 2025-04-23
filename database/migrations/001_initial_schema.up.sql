-- +migrate Up
-- SQL in this section is executed when the migration is applied

-- Create users table
CREATE TABLE IF NOT EXISTS users (
    id SERIAL PRIMARY KEY,
    username TEXT UNIQUE NOT NULL,
    email TEXT UNIQUE NOT NULL,
    password TEXT NOT NULL,
    name TEXT NOT NULL,
    photo_url TEXT,
    role TEXT NOT NULL DEFAULT 'user',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Create locations table (for cities/regions)
CREATE TABLE IF NOT EXISTS locations (
    id SERIAL PRIMARY KEY,
    name TEXT UNIQUE NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Create cctvs table
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

-- Create indexes
CREATE INDEX IF NOT EXISTS idx_cctvs_location_id ON cctvs(location_id);
CREATE INDEX IF NOT EXISTS idx_cctvs_is_active ON cctvs(is_active);

-- Insert initial data
INSERT INTO locations (name) VALUES 
('Jakarta'), ('Bogor'), ('Depok'), ('Tangerang'), ('Bekasi'),
('Bandung'), ('Semarang'), ('Yogyakarta'), ('Surabaya'), ('Medan'),
('Makassar'), ('Palembang'), ('Balikpapan'), ('Batam'), ('Pekanbaru')
ON CONFLICT (name) DO NOTHING;

-- Insert admin user (password: admin123)
INSERT INTO users (username, email, password, name, role) VALUES
('admin', 'admin@cctv.app', '$2a$10$X8z5sZ5JZ5Z5Z5Z5Z5Z5.e5Z5Z5Z5Z5Z5Z5Z5Z5Z5Z5Z5Z5Z5', 'Admin', 'admin')
ON CONFLICT (username) DO NOTHING;

-- +migrate Down
-- SQL in this section is executed when the migration is rolled back
DROP INDEX IF EXISTS idx_cctvs_location_id;
DROP INDEX IF EXISTS idx_cctvs_is_active;
DROP TABLE IF EXISTS cctvs;
DROP TABLE IF EXISTS locations;
DROP TABLE IF EXISTS users;