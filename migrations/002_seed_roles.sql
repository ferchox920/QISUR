INSERT INTO roles (name)
VALUES ('admin'), ('user'), ('client')
ON CONFLICT DO NOTHING;
