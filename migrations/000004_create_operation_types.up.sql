-- Создание таблицы типов операций

CREATE TABLE IF NOT EXISTS operation_types (
    id SERIAL PRIMARY KEY,
    code VARCHAR(50) UNIQUE NOT NULL,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    checklist_template TEXT,
    is_active BOOLEAN DEFAULT true NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Индексы
CREATE INDEX idx_operation_types_code ON operation_types(code);
CREATE INDEX idx_operation_types_is_active ON operation_types(is_active);

-- Заполняем начальными данными
INSERT INTO operation_types (code, name, description, checklist_template, is_active) VALUES
('PHACOEMULSIFICATION', 'Факоэмульсификация катаракты', 'Хирургическое удаление катаракты с имплантацией интраокулярной линзы', '[]', true),
('ANTIGLAUCOMA', 'Антиглаукоматозная операция', 'Хирургическое лечение глаукомы для снижения внутриглазного давления', '[]', true),
('VITRECTOMY', 'Витрэктомия', 'Хирургическое удаление стекловидного тела глаза', '[]', true)
ON CONFLICT (code) DO NOTHING;
