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
}

type Store interface {
	Search() ([]Project, error)
	FetchByID(id int) (Project, error)
	Create(name string, author auth.User) (int, error)
	Update(p Project) error
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
	q := "SELECT project.id, project.name, project.modtime, project.created, account.id, account.login FROM project INNER JOIN account ON project.author_id = account.id;"

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

		err := rows.Scan(&p.ID, &p.Name, &p.Modtime, &p.Created, &u.ID, &u.Name)
		if p.Modtime != nil {
			p.ModtimeUTS = p.Modtime.Unix()
		}
		p.CreatedUTS = p.Created.Unix()
		if err != nil {
			return ps, err
		}

		p.Author = &u

		ps = append(ps, p)
	}

	return ps, rows.Err()
}

func (s PostgresStore) FetchByID(id int) (Project, error) {
	q := "SELECT id, name, author_id, modtime, created FROM project WHERE id=$1;"

	p := Project{}
	var uid int

	row := s.DB.QueryRow(q, id)
	err := row.Scan(&p.ID, &p.Name, &uid, &p.Modtime, &p.Created)
	p.Author = &auth.User{ID: uint(uid)}

	if err == sql.ErrNoRows {
		return p, ErrNoFound
	}

	return p, err
}

func (s PostgresStore) Create(name string, author auth.User) (int, error) {
	// TODO: Make these prepared statements.

	var id int
	err := s.DB.QueryRow(`INSERT INTO project (name, author_id) VALUES ($1, $2) RETURNING id;`, name, author.ID).Scan(&id)

	return id, err
}

func (s PostgresStore) Update(p Project) error {
	q := `UPDATE project SET name = $2, author_id = $3, modtime = $4 WHERE id = $1;`

	_, err := s.DB.Exec(q, p.ID, p.Name, p.Author.ID, time.Now())
	if err != nil {
		pqErr := err.(*pq.Error)
		if pqErr.Code == "23503" {
			return ErrInvalidAuthor
		}
	}

	return err
}
