ALTER TABLE project ADD PRIMARY KEY (id);

CREATE TABLE revision (
	id SERIAL,
	content TEXT,
	project_id integer REFERENCES project,
	created timestamp DEFAULT current_timestamp,

	constraint fk_project_revision foreign key (project_id) REFERENCES project (id),

	PRIMARY KEY (id)
);

INSERT INTO revision (content, project_id) VALUES ('Hoi piepeloi', 1);
