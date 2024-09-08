CREATE TABLE users (
  id UUID PRIMARY KEY NOT NULL
);

CREATE TABLE identity_users (
    id TEXT PRIMARY KEY NOT NULL,
    user_id UUID NOT NULL REFERENCES users(id)
);

CREATE TABLE hibp_cases (
  id UUID PRIMARY KEY NOT NULL,
  user_id UUID NOT NULL REFERENCES users(id),
  name TEXT NOT NULL
);

INSERT INTO users
SELECT MD5(CONCAT('user',series))::UUID AS id
FROM GENERATE_SERIES(0,10,1) AS series;

INSERT INTO identity_users
SELECT 
  CONCAT('email|', LEFT(REPLACE(MD5(CONCAT('identity',series))::VARCHAR, '-', ''), 24)) AS id,
  MD5(CONCAT('user',series))::UUID AS user_id
FROM GENERATE_SERIES(0,10,1) AS series;

INSERT INTO hibp_cases
SELECT 
  MD5(CONCAT('case',series))::UUID AS id,
  MD5(CONCAT('user',series))::UUID AS user_id,
  'Adobe' AS name
FROM GENERATE_SERIES(0,10,1) AS series;