INSERT INTO permissions(code,description) VALUES
('roles:read','Read roles'),('roles:write','Manage roles'),
('patients:read','Read patient directory'),('patients:write','Manage patients'),
('appointments:read','Read appointments'),('appointments:write','Manage appointments'),
('finance:read','Read finance'),('finance:write','Manage finance')
ON CONFLICT(code) DO NOTHING;
