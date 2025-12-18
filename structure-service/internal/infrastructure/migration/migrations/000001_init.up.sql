-- Создание таблицы universities (вузы)
CREATE TABLE IF NOT EXISTS universities (
  id SERIAL PRIMARY KEY,
  name TEXT NOT NULL,
  inn TEXT NOT NULL,
  kpp TEXT,
  foiv TEXT,
  created_at TIMESTAMP WITH TIME ZONE DEFAULT now(),
  updated_at TIMESTAMP WITH TIME ZONE DEFAULT now(),
  UNIQUE(inn, kpp)
);

-- Создание индексов для universities
CREATE INDEX IF NOT EXISTS idx_universities_inn ON universities(inn);
CREATE INDEX IF NOT EXISTS idx_universities_name ON universities(name);

-- Создание таблицы branches (филиалы/подразделения)
CREATE TABLE IF NOT EXISTS branches (
  id SERIAL PRIMARY KEY,
  university_id INTEGER NOT NULL REFERENCES universities(id) ON DELETE CASCADE,
  name TEXT NOT NULL,
  created_at TIMESTAMP WITH TIME ZONE DEFAULT now(),
  updated_at TIMESTAMP WITH TIME ZONE DEFAULT now()
);

-- Создание индексов для branches
CREATE INDEX IF NOT EXISTS idx_branches_university_id ON branches(university_id);
CREATE INDEX IF NOT EXISTS idx_branches_name ON branches(name);

-- Создание таблицы faculties (факультеты/институты)
CREATE TABLE IF NOT EXISTS faculties (
  id SERIAL PRIMARY KEY,
  branch_id INTEGER REFERENCES branches(id) ON DELETE CASCADE,
  name TEXT NOT NULL,
  created_at TIMESTAMP WITH TIME ZONE DEFAULT now(),
  updated_at TIMESTAMP WITH TIME ZONE DEFAULT now()
);

-- Создание индексов для faculties
CREATE INDEX IF NOT EXISTS idx_faculties_branch_id ON faculties(branch_id);
CREATE INDEX IF NOT EXISTS idx_faculties_name ON faculties(name);

-- Создание таблицы groups (группы)
CREATE TABLE IF NOT EXISTS groups (
  id SERIAL PRIMARY KEY,
  faculty_id INTEGER NOT NULL REFERENCES faculties(id) ON DELETE CASCADE,
  course INTEGER NOT NULL,
  number TEXT NOT NULL,
  chat_id INTEGER, -- ID чата из chat-service (может быть NULL)
  created_at TIMESTAMP WITH TIME ZONE DEFAULT now(),
  updated_at TIMESTAMP WITH TIME ZONE DEFAULT now()
);

-- Создание индексов для groups
CREATE INDEX IF NOT EXISTS idx_groups_faculty_id ON groups(faculty_id);
CREATE INDEX IF NOT EXISTS idx_groups_chat_id ON groups(chat_id);
CREATE INDEX IF NOT EXISTS idx_groups_course ON groups(course);

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

CREATE TRIGGER update_branches_updated_at BEFORE UPDATE ON branches
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_faculties_updated_at BEFORE UPDATE ON faculties
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_groups_updated_at BEFORE UPDATE ON groups
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

