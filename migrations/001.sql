ALTER TABLE account ADD PRIMARY KEY (id);

CREATE TABLE project (
	id SERIAL,
	name VARCHAR(255) NOT NULL,
	author_id integer REFERENCES account,
	modtime timestamp,
	created timestamp DEFAULT current_timestamp,

	constraint fk_project_account foreign key (author_id) REFERENCES account (id)
);

INSERT INTO project (name, author_id) VALUES ('Example project', 1);
