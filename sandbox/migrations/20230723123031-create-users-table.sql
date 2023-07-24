-- Up
BEGIN

CREATE TABLE users (
  user_id INT,
  last_name VARCHAR(255),
  first_name VARCHAR(255),
  created_at TIMESTAMPTZ 
);

END


-- Down
BEGIN

DROP TABLE users;

END
