-- Создание таблицы universities (вузы)
CREATE TABLE IF NOT EXISTS universities (
  id SERIAL PRIMARY KEY,
  name TEXT NOT NULL,
  inn TEXT NOT NULL,
  kpp TEXT,
  created_at TIMESTAMP WITH TIME ZONE DEFAULT now(),
  updated_at TIMESTAMP WITH TIME ZONE DEFAULT now(),
  UNIQUE(inn, kpp)
);

-- Создание индексов для universities
CREATE INDEX IF NOT EXISTS idx_universities_inn ON universities(inn);
CREATE INDEX IF NOT EXISTS idx_universities_name ON universities(name);

-- Создание таблицы employees (сотрудники)
CREATE TABLE IF NOT EXISTS employees (
  id SERIAL PRIMARY KEY,
  first_name TEXT NOT NULL,
  last_name TEXT NOT NULL,
  middle_name TEXT,
  phone TEXT NOT NULL,
  max_id TEXT,
  inn TEXT,
  kpp TEXT,
  university_id INTEGER NOT NULL REFERENCES universities(id) ON DELETE CASCADE,
  created_at TIMESTAMP WITH TIME ZONE DEFAULT now(),
  updated_at TIMESTAMP WITH TIME ZONE DEFAULT now()
);

-- Создание индексов для employees
CREATE INDEX IF NOT EXISTS idx_employees_phone ON employees(phone);
CREATE INDEX IF NOT EXISTS idx_employees_max_id ON employees(max_id);
CREATE INDEX IF NOT EXISTS idx_employees_university_id ON employees(university_id);
CREATE INDEX IF NOT EXISTS idx_employees_name ON employees(last_name, first_name);
CREATE INDEX IF NOT EXISTS idx_employees_inn ON employees(inn);

-- Триггер для автоматического обновления updated_at
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = now();
    RETURN NEW;
END;
$$ language 'plpgsql';

CREATE TRIGGER update_universities_updated_at BEFORE UPDATE ON universities
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_employees_updated_at BEFORE UPDATE ON employees
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

