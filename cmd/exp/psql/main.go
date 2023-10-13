package main

import (
	"database/sql"
	"fmt"

	_ "github.com/jackc/pgx/v4/stdlib"
)

type PostgresConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	Database string
	SSLMode  string
}

func (cfg PostgresConfig) String() string {
	return fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s", cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.Database, cfg.SSLMode)
}

func main() {
	cfg := PostgresConfig{
		Host:     "localhost",
		Port:     "5432",
		User:     "sa",
		Password: "@dmin1234",
		Database: "lenslocked",
		SSLMode:  "disable",
	}

	db, err := sql.Open("pgx", cfg.String())
	if err != nil {
		panic(err)
	}
	defer db.Close()
	err = db.Ping()
	if err != nil {
		panic(err)
	}

	fmt.Println("Connected!")

	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS users (
			id SERIAL PRIMARY KEY,
			first_name TEXT,
  			last_name TEXT,
			email TEXT UNIQUE NOT NULL
		);

		CREATE TABLE IF NOT EXISTS orders (
			id SERIAL PRIMARY KEY,
			user_id INT NOT NULL,
			amount INT,
			description TEXT,
			FOREIGN KEY (user_id) REFERENCES users (id)
		);`)
	if err != nil {
		panic(err)
	}
	fmt.Println("Tables created.")

	first_name := "User"
	last_name := "Test"
	email := "user@test.com"
	row := db.QueryRow(`INSERT INTO users (first_name, last_name, email) VALUES ($1, $2, $3) RETURNING id;`, first_name, last_name, email)
	var id int
	err = row.Scan(&id)
	if err != nil {
		panic(err)
	}
	fmt.Println("User created. id =", id)

	id = 1
	row = db.QueryRow(`
		SELECT first_name, last_name, email
		FROM users
		WHERE id=$1;`, id)
	err = row.Scan(&first_name, &last_name, &email)
	if err != nil {
		panic(err)
	}
	fmt.Printf("User information: first_name=%s, last_name=%s, email=%s\n", first_name, last_name, email)

	for i := 1; i <= 5; i++ {
		amount := i * 100
		desc := fmt.Sprintf("Fake order #%d", i)
		_, err := db.Exec(`
		INSERT INTO orders(user_id, amount, description)
		VALUES($1, $2, $3)`, id, amount, desc)
		if err != nil {
			panic(err)
		}
	}
	fmt.Println("Created fake orders.")

	type Order struct {
		ID          int
		UserID      int
		Amount      int
		Description string
	}

	var orders []Order

	rows, err := db.Query(`
		SELECT id, amount, description
		FROM orders
		WHERE user_id=$1`, id)
	if err != nil {
		panic(err)
	}
	defer rows.Close()

	for rows.Next() {
		var order Order
		order.UserID = id
		err := rows.Scan(&order.ID, &order.Amount, &order.Description)
		if err != nil {
			panic(err)
		}
		orders = append(orders, order)
	}
	err = rows.Err()
	if err != nil {
		panic(err)
	}

	fmt.Println("Orders:", orders)
}
