
BEGIN -- UP

CREATE TABLE users (
  id          INT               NOT NULL  IDENTITY  PRIMARY KEY,
  last_name   VARCHAR(255)      NOT NULL,
  first_name  VARCHAR(255),
  created_at  DATETIMEOFFSET(7)
);

END -- UP


BEGIN -- DOWN

DROP TABLE users;

END -- DOWN
