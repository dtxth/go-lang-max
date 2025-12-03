-- Создание таблицы department_managers (операторы, назначенные на подразделения)
CREATE TABLE IF NOT EXISTS department_managers (
  id SERIAL PRIMARY KEY,
  employee_id INTEGER NOT NULL, -- Reference to employee-service
  branch_id INTEGER REFERENCES branches(id) ON DELETE CASCADE,
  faculty_id INTEGER REFERENCES faculties(id) ON DELETE CASCADE,
  assigned_by INTEGER, -- Reference to curator user_id
  assigned_at TIMESTAMP WITH TIME ZONE DEFAULT now(),
  UNIQUE(employee_id, branch_id, faculty_id),
  CHECK (branch_id IS NOT NULL OR faculty_id IS NOT NULL)
);

-- Создание индексов для department_managers
CREATE INDEX IF NOT EXISTS idx_department_managers_employee_id ON department_managers(employee_id);
CREATE INDEX IF NOT EXISTS idx_department_managers_branch_id ON department_managers(branch_id);
CREATE INDEX IF NOT EXISTS idx_department_managers_faculty_id ON department_managers(faculty_id);
