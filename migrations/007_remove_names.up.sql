ALTER TABLE members DROP COLUMN first_name;
ALTER TABLE members DROP COLUMN middle_name;
ALTER TABLE members DROP COLUMN last_name;
ALTER TABLE members DROP COLUMN nickname;
ALTER TABLE members ADD COLUMN name TEXT NOT NULL DEFAULT "";