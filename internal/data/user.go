package data

import (
	"database/sql"
	"errors"
	"time"

	"github.com/AbrahamMayowa/ticketmania/internal/validator"
	"github.com/lib/pq"
	"golang.org/x/crypto/bcrypt"
)

type User struct {
	Id int64 `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	Email string `json:"email"`
	Password password `json:"-"`
	Version int32 `json:"version"`
}

type password struct {
	Plaintext *string
	Hash      []byte
}


type UserModel struct {
	DB *sql.DB
}



func (p *password) Set(plaintextPassword string) error {

	hash, err := bcrypt.GenerateFromPassword([]byte(plaintextPassword), 12)
	if err != nil {
		return err
	}
	p.Plaintext = &plaintextPassword
	p.Hash = hash
	return nil
}

func (p *password) Matches(plaintextPassword string) (bool, error) {
	err := bcrypt.CompareHashAndPassword(p.Hash, []byte(plaintextPassword))
	if err != nil {
		if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}


func (u UserModel) Insert(user *User) error {
	userData := []interface{}{user.Email, user.Password.Hash}
	query := `INSERT INTO users (email, password) VALUES ($1, $2) RETURNING id, created_at, email, version`
	err := u.DB.QueryRow(query, userData...).Scan(&user.Id, &user.CreatedAt, &user.Email, &user.Version)
	if err != nil {
	if pgErr, ok := err.(*pq.Error); ok {
			if pgErr.Code == "23505" {
				return ErrUserAlreadyExists
			}
		}
		return err
	}
	return nil
}

func (u UserModel) GetByEmail(email string) (*User, error) {
	query := `SELECT id, created_at, email, password, version FROM users WHERE email = $1`
	var user User
	err := u.DB.QueryRow(query, email).Scan(&user.Id, &user.CreatedAt, &user.Email, &user.Password.Hash, &user.Version)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrRecordNotFound
		}
		return nil, err
	}
	return &user, nil
}

func ValidateUser(v *validator.Validator, user *User) {
	v.Check(validator.Matches(user.Email, validator.EmailRegex), "email", "must be a valid email address")
	v.Check(user.Password.Plaintext != nil, "password", "must be provided")
}





