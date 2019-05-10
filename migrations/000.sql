CREATE TABLE account (
	id SERIAL,
	login VARCHAR(30) UNIQUE NOT NULL,
	password VARCHAR(255) NOT NULL,
	modtime timestamp,
	created timestamp DEFAULT current_timestamp
);

INSERT INTO account (login, password) VALUES ('example', '$2a$10$7iU90lWUss3yuvx8q4cg2.vBMaJkDpHsQCeRL1FZhLkCSFtWWkkzu');
/* password: example */
