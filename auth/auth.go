package auth

import (
	"database/sql"
	"errors"

	"golang.org/x/crypto/bcrypt"
)

type Store interface {
	CheckLogin(name string, password string) (User, error)
	Register(name string, password string) error
}

type PostgresStore struct {
	DB *sql.DB
}

type User struct {
	ID   uint   `json:"id"`
	Name string `json:"name"`
}

var (
	ErrNoFound       = errors.New("no users found")
	ErrAlreadyExists = errors.New("user already exists")
)

func (s PostgresStore) CheckLogin(name string, password string) (User, error) {
	u := User{}

	row := s.DB.QueryRow(`SELECT id, login, password FROM account WHERE login=$1;`,
		name)
	var passwd string
	err := row.Scan(&u.ID, &u.Name, &passwd)
	if err == sql.ErrNoRows {
		return u, ErrNoFound
	}
	if err != nil {
		return u, err
	}

	err = bcrypt.CompareHashAndPassword([]byte(passwd), []byte(password))

	return u, err
}

func (s PostgresStore) Register(name string, password string) error {
	// TODO: Make these prepared statements.

	passwd, err := bcrypt.GenerateFromPassword([]byte(password), -1)
	if err != nil {
		return err
	}

	result, err := s.DB.Exec(`INSERT INTO account (login, password) VALUES ($1, $2);`,
		name, passwd)
	if err != nil {
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows != 1 {
		return ErrAlreadyExists
	}

	return err
}
