-- Create roles table
CREATE TABLE IF NOT EXISTS roles (
  id SERIAL PRIMARY KEY,
  name TEXT NOT NULL UNIQUE,
  description TEXT,
  created_at TIMESTAMP WITH TIME ZONE DEFAULT now()
);

-- Insert default roles
INSERT INTO roles (name, description) VALUES
  ('superadmin', 'Представитель VK с полными правами'),
  ('curator', 'Ответственный представитель от вуза'),
  ('operator', 'Представитель подразделения вуза')
ON CONFLICT (name) DO NOTHING;

-- Create user_roles table with context
CREATE TABLE IF NOT EXISTS user_roles (
  id SERIAL PRIMARY KEY,
  user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  role_id INTEGER NOT NULL REFERENCES roles(id) ON DELETE CASCADE,
  university_id INTEGER,
  branch_id INTEGER,
  faculty_id INTEGER,
  assigned_by INTEGER REFERENCES users(id),
  assigned_at TIMESTAMP WITH TIME ZONE DEFAULT now(),
  UNIQUE(user_id, role_id, university_id, branch_id, faculty_id)
);

-- Create indexes for efficient querying
CREATE INDEX IF NOT EXISTS idx_user_roles_user_id ON user_roles(user_id);
CREATE INDEX IF NOT EXISTS idx_user_roles_role_id ON user_roles(role_id);
CREATE INDEX IF NOT EXISTS idx_user_roles_university_id ON user_roles(university_id);

-- Add comments
COMMENT ON TABLE roles IS 'Роли пользователей в системе';
COMMENT ON TABLE user_roles IS 'Связь пользователей с ролями и контекстом (вуз, филиал, факультет)';
COMMENT ON COLUMN user_roles.university_id IS 'NULL для superadmin, заполнено для curator и operator';
COMMENT ON COLUMN user_roles.branch_id IS 'NULL для curator, может быть заполнено для operator';
COMMENT ON COLUMN user_roles.faculty_id IS 'NULL для curator, может быть заполнено для operator';
