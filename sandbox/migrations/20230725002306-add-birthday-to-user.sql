
BEGIN -- UP

ALTER TABLE users ADD active BIT;
UPDATE users SET active = 1;
ALTER TABLE users
    ALTER COLUMN active BIT NOT NULL;

END -- UP


BEGIN -- DOWN

ALTER TABLE users
    DROP COLUMN active;

END -- DOWN
