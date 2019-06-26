package project

import (
	"database/sql"
	"errors"
	"time"

	"github.com/lib/pq"
)

type Revision struct {
	ID      *int    `json:"id"`
	Content *string `json:"content"`

	Created    *time.Time `json:"-"`
	CreatedUTS int64      `json:"created"`
}

var (
	ErrInvalidProject = errors.New("invalid project")
)

func (s PostgresStore) SaveRevision(pid int, content string) error {
	_, err := s.DB.Exec(`INSERT INTO revision (content, project_id) VALUES ($1, $2);`, content, pid)
	if err != nil {
		pqErr := err.(*pq.Error)
		if pqErr.Code == "23503" {
			return ErrInvalidProject
		}
	}
	return err
}

func (s PostgresStore) FetchLatestRevisionByProject(pid int) (Revision, error) {
	q := "SELECT id, content, created FROM revision WHERE project_id=$1 ORDER BY created DESC LIMIT 1;"

	row := s.DB.QueryRow(q, pid)

	r := Revision{}

	err := row.Scan(&r.ID, &r.Content, &r.Created)
	if err == sql.ErrNoRows {
		return r, nil
	}
	r.CreatedUTS = r.Created.Unix()

	return r, err
}
