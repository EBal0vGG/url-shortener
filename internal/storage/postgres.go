package storage

import (
	"database/sql"
	"errors"

	_ "github.com/lib/pq"
)

// Postgres — обёртка для *sql.DB
type Postgres struct {
	DB *sql.DB
}

// NewPostgres создаёт обёртку
func NewPostgres(db *sql.DB) *Postgres {
	return &Postgres{DB: db}
}

// Save сохраняет короткую и длинную ссылку.
// Если короткая ссылка уже существует, ничего не делает.
func (p *Postgres) Save(short, original string) error {
	_, err := p.DB.Exec(
		`INSERT INTO urls (short, original) VALUES ($1, $2) ON CONFLICT (short) DO NOTHING`,
		short, original,
	)
	return err
}

// Find получает длинную ссылку по короткой.
func (p *Postgres) Find(short string) (string, error) {
	var original string
	err := p.DB.QueryRow(`SELECT original FROM urls WHERE short = $1`, short).Scan(&original)
	if err == sql.ErrNoRows {
		return "", errors.New("not found")
	}
	return original, err
}

// IncrementClicks увеличивает счетчик переходов.
func (p *Postgres) IncrementClicks(short string) error {
	_, err := p.DB.Exec(`UPDATE urls SET clicks = clicks + 1 WHERE short = $1`, short)
	return err
}

// GetClicks возвращает количество переходов по короткой ссылке.
func (p *Postgres) GetClicks(short string) (int, error) {
	var clicks int
	err := p.DB.QueryRow(`SELECT clicks FROM urls WHERE short = $1`, short).Scan(&clicks)
	if err != nil {
		return 0, err
	}
	return clicks, nil
}
