package project

import (
	"database/sql"
	"errors"
	"time"

	"github.com/frengine/server/auth"
	"github.com/lib/pq"
)

type Project struct {
	ID         int        `json:"id"`
	Name       string     `json:"name"`
	Author     *auth.User `json:"author,omitempty"`
	Modtime    *time.Time `json:"-"`
	ModtimeUTS int64      `json:"modtime"`
	Created    *time.Time `json:"-"`
	CreatedUTS int64      `json:"created"`

	Revision *Revision `json:"revision"`

	TouchedUTS int64 `json:"touched"`

	Deleted *time.Time `json:"-"`
}

func (p Project) LastModified() time.Time {
	if p.Modtime != nil {
		return *p.Modtime
	}

	if p.Created != nil {
		return *p.Created
	}

	return time.Time{}
}

type Store interface {
	Search() ([]Project, error)
	FetchByID(id int) (Project, error)
	Create(name string, author auth.User) (int, error)
	Update(p Project) error
	Delete(id int) error

	SaveRevision(pid int, content string) error
	FetchLatestRevisionByProject(pid int) (Revision, error)
}

type PostgresStore struct {
	DB *sql.DB
}

var (
	ErrNoFound       = errors.New("no users found")
	ErrAlreadyExists = errors.New("user already exists")
	ErrInvalidAuthor = errors.New("invalid author")
)

func (s PostgresStore) Search() ([]Project, error) {
	q := `
	SELECT project.id, project.name, project.modtime, project.created, account.id, account.login, revision.id, revision.content, revision.created
	FROM project
	INNER JOIN account
		ON project.author_id = account.id
	LEFT JOIN revision
		ON revision.id = (SELECT id FROM revision WHERE project_id = project.id ORDER BY created DESC LIMIT 1)
	WHERE project.deleted IS NULL`

	rows, err := s.DB.Query(q)
	if err != nil {
		if err == sql.ErrNoRows {
			return []Project{}, ErrNoFound
		}
		return []Project{}, err
	}

	ps := []Project{}

	for rows.Next() {
		p := Project{}
		u := auth.User{}
		r := Revision{}

		err := rows.Scan(&p.ID, &p.Name, &p.Modtime, &p.Created, &u.ID, &u.Name, &r.ID, &r.Content, &r.Created)
		if err != nil {
			return ps, err
		}

		if p.Modtime != nil {
			p.ModtimeUTS = p.Modtime.Unix()
		}
		if p.Created != nil {
			p.CreatedUTS = p.Created.Unix()
		}
		if r.Created != nil {
			r.CreatedUTS = r.Created.Unix()
		}

		p.CreatedUTS = p.Created.Unix()
		p.TouchedUTS = max(p.ModtimeUTS, r.CreatedUTS, p.CreatedUTS)

		p.Author = &u

		p.Revision = &r

		ps = append(ps, p)
	}

	return ps, rows.Err()
}

func (s PostgresStore) FetchByID(id int) (Project, error) {
	q := `SELECT project.id, project.name, project.modtime, project.created, account.id, account.login, revision.id, revision.content, revision.created
	FROM project
	INNER JOIN account
		ON project.author_id = account.id
	LEFT JOIN revision
		ON revision.id = (SELECT id FROM revision WHERE project_id = project.id ORDER BY created DESC LIMIT 1)
	WHERE project.id=$1 AND deleted IS NULL;`

	p := Project{}
	u := auth.User{}
	r := Revision{}

	row := s.DB.QueryRow(q, id)

	err := row.Scan(&p.ID, &p.Name, &p.Modtime, &p.Created, &u.ID, &u.Name, &r.ID, &r.Content, &r.Created)
	if err != nil {
		if err == sql.ErrNoRows {
			return p, ErrNoFound
		}
		return p, err
	}

	if p.Modtime != nil {
		p.ModtimeUTS = p.Modtime.Unix()
	}
	if p.Created != nil {
		p.CreatedUTS = p.Created.Unix()
	}
	if r.Created != nil {
		r.CreatedUTS = r.Created.Unix()
	}

	p.CreatedUTS = p.Created.Unix()
	p.TouchedUTS = max(p.ModtimeUTS, r.CreatedUTS, p.CreatedUTS)

	p.Author = &u

	p.Revision = &r

	return p, err
}

func max(nums ...int64) int64 {
	var max int64

	for _, num := range nums {
		if num > max {
			max = num
		}
	}

	return max
}

func (s PostgresStore) Create(name string, author auth.User) (int, error) {
	// TODO: Make these prepared statements.

	var id int
	err := s.DB.QueryRow(`INSERT INTO project (name, author_id) VALUES ($1, $2) RETURNING id;`, name, author.ID).Scan(&id)

	return id, err
}

func (s PostgresStore) Update(p Project) error {
	q := `UPDATE project SET name = $2, author_id = $3, modtime = NOW() WHERE id = $1;`

	_, err := s.DB.Exec(q, p.ID, p.Name, p.Author.ID)
	if err != nil {
		pqErr := err.(*pq.Error)
		if pqErr.Code == "23503" {
			return ErrInvalidAuthor
		}
	}

	return err
}

func (s PostgresStore) Delete(id int) error {
	q := `UPDATE project SET deleted = NOW() WHERE id = $1;`

	_, err := s.DB.Exec(q, id)
	return err
}
