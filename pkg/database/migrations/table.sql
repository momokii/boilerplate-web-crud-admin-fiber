CREATE TABLE users (
    id SERIAL NOT NULL PRIMARY KEY,
    username VARCHAR(255) NOT NULL UNIQUE,
    password VARCHAR(255) NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NULL DEFAULT CURRENT_TIMESTAMP,
    role BIGINT NULL DEFAULT 2,
    is_deleted BOOLEAN DEFAULT FALSE,
    CONSTRAINT users_role_fkey FOREIGN KEY (role) REFERENCES role(id)
);

CREATE TABLE role (
    id BIGSERIAL NOT NULL PRIMARY KEY,
    name VARCHAR(255) NOT NULL UNIQUE
);

CREATE TABLE project_status (
    id BIGSERIAL NOT NULL PRIMARY KEY,
    name VARCHAR(255) NOT NULL UNIQUE
);

CREATE TABLE projects (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    start_date DATE,
    end_date DATE,
    status INT,
    budget BIGINT DEFAULT 0,
    created_by INT NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (status) REFERENCES project_status(id),
    FOREIGN KEY (created_by) REFERENCES users(id) ON DELETE CASCADE
);

CREATE TABLE daily_logs (
    id SERIAL PRIMARY KEY,
    project_id INT,
    log_date DATE,
    description TEXT,
    issues TEXT,
    income BIGINT DEFAULT 0,
    expense BIGINT DEFAULT 0,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    file TEXT DEFAULT NULL,
    FOREIGN KEY (project_id) REFERENCES projects(id) ON DELETE CASCADE
);

--  BELOW IS NOT IMPLEMENTED YET
-- CREATE TABLE task_status (
--     id SERIAL PRIMARY KEY,
--     name VARCHAR(255) NOT NULL UNIQUE
-- );

-- CREATE TABLE tasks (
--     id SERIAL PRIMARY KEY,
--     project_id INT,
--     name VARCHAR(255) NOT NULL,
--     description TEXT,
--     start_date DATE,
--     end_date DATE,
--     status INT,
--     created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
--     updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
--     FOREIGN KEY (status) REFERENCES task_status(id),
--     FOREIGN KEY (project_id) REFERENCES projects(id) ON DELETE CASCADE
-- );
