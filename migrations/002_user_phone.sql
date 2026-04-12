-- For databases created before `phone` existed on `users`.
ALTER TABLE users ADD COLUMN IF NOT EXISTS phone TEXT;
