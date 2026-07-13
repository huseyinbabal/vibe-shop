-- Identity moves to Keycloak: user_id now stores the verified token's sub
-- claim (a string), and the local users table goes away with slice 5.
ALTER TABLE cart DROP CONSTRAINT cart_user_id_fkey;
ALTER TABLE orders DROP CONSTRAINT orders_user_id_fkey;
ALTER TABLE cart ALTER COLUMN user_id TYPE TEXT USING user_id::text;
ALTER TABLE orders ALTER COLUMN user_id TYPE TEXT USING user_id::text;
DROP TABLE users;
