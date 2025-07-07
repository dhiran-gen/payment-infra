package internal

import (
	"database/sql"

	_ "github.com/lib/pq"
)

type Payment struct {
	ID     string
	Amount int
}

type DB struct {
	*sql.DB
}

func NewDB(dsn string) (*DB, error) {
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, err
	}
	return &DB{db}, nil
}

func (db *DB) CreatePayment(p Payment) error {
	_, err := db.Exec(`INSERT INTO payments (id, amount) VALUES ($1, $2)`, p.ID, p.Amount)
	return err
}

func (db *DB) GetPayment(id string) (*Payment, error) {
	row := db.QueryRow(`SELECT id, amount FROM payments WHERE id = $1`, id)
	var p Payment
	if err := row.Scan(&p.ID, &p.Amount); err != nil {
		return nil, err
	}
	return &p, nil
}

func (db *DB) ListPayments() ([]Payment, error) {
	rows, err := db.Query(`SELECT id, amount FROM payments`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var payments []Payment
	for rows.Next() {
		var p Payment
		if err := rows.Scan(&p.ID, &p.Amount); err != nil {
			return nil, err
		}
		payments = append(payments, p)
	}
	return payments, nil
}
