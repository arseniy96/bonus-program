BEGIN TRANSACTION;

    CREATE TABLE IF NOT EXISTS users(
        id INT PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
        login VARCHAR,
        password VARCHAR,
        cookie VARCHAR,
        cookie_expired TIME
    );

    CREATE UNIQUE INDEX IF NOT EXISTS login_idx on users(login);

COMMIT;