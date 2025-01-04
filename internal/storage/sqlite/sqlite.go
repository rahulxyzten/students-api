package sqlite

import (
	"database/sql"
	"fmt"

	"github.com/rahulxyzten/students-api/internal/config"
	"github.com/rahulxyzten/students-api/internal/types"

	_ "github.com/mattn/go-sqlite3"
)

type Sqlite struct {
	Db *sql.DB
}

func New(cfg *config.Config) (*Sqlite, error) {
	db, err := sql.Open("sqlite3", cfg.StoragePath)
	if err != nil {
		return nil, err
	}

	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS students(
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	name TEXT,
	email TEXT,
	age INTEGER
	)`)

	if err != nil {
		return nil, err
	}

	return &Sqlite{
		Db: db,
	}, nil
}

func (s *Sqlite) CreateStudent(name string, email string, age int) (int64, error) {

	stmt, err := s.Db.Prepare("INSERT INTO students (name, email, age) VALUES (?, ?, ?)")
	if err != nil {
		return 0, err
	}

	defer stmt.Close()

	result, err := stmt.Exec(name, email, age)
	if err != nil {
		return 0, err
	}

	lastId, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}

	return lastId, nil
}

func (s *Sqlite) GetStudentById(id int64) (types.Student, error) {

	stmt, err := s.Db.Prepare("SELECT id, name, email, age FROM students WHERE id = ? LIMIT 1")
	if err != nil {
		return types.Student{}, err
	}

	defer stmt.Close()

	var student types.Student

	err = stmt.QueryRow(id).Scan(&student.Id, &student.Name, &student.Email, &student.Age)
	if err != nil {
		if err == sql.ErrNoRows {
			return types.Student{}, fmt.Errorf("not student found with id %s", fmt.Sprint(id))
		}
		return types.Student{}, fmt.Errorf("query error: %w", err)
	}

	return student, nil
}

func (s *Sqlite) GetStudents() ([]types.Student, error) {
	stmt, err := s.Db.Prepare("SELECT id, name, email, age FROM students")
	if err != nil {
		return nil, err
	}

	defer stmt.Close()

	rows, err := stmt.Query()
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var students []types.Student

	for rows.Next() {
		var student types.Student

		err := rows.Scan(&student.Id, &student.Name, &student.Email, &student.Age)
		if err != nil {
			return nil, err
		}

		students = append(students, student)
	}

	return students, nil
}

func (s *Sqlite) DeleteStudentById(id int64) error {
	stmt, err := s.Db.Prepare("DELETE FROM students WHERE id = ?")
	if err != nil {
		return err
	}

	defer stmt.Close()

	result, err := stmt.Exec(id)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return fmt.Errorf("no student found with id %s", fmt.Sprint(id))
	}

	return nil
}

func (s *Sqlite) UpdateStudentById(id int64, updates map[string]interface{}) (types.Student, error) {
	query := "UPDATE students SET"
	args := []interface{}{}
	//Initialize a slice of interface{}
	counter := 0

	if name, ok := updates["name"]; ok {
		query += " name = ?"
		args = append(args, name)
		counter++
	}
	// in the name variable it assign the value associated with key name

	if email, ok := updates["email"]; ok {
		if counter > 0 {
			query += ","
		}
		query += " email = ?"
		args = append(args, email)
		counter++
	}

	if age, ok := updates["age"]; ok {
		if counter > 0 {
			query += ","
		}
		query += " age = ?"
		args = append(args, age)
	}

	if counter == 0 {
		return types.Student{}, fmt.Errorf("no fields to update")
	}

	query += " WHERE id = ?"
	args = append(args, id)

	stmt, err := s.Db.Prepare(query)
	if err != nil {
		return types.Student{}, err
	}
	defer stmt.Close()

	result, err := stmt.Exec(args...)
	if err != nil {
		return types.Student{}, err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return types.Student{}, err
	}

	if rowsAffected == 0 {
		return types.Student{}, fmt.Errorf("no student found with id %s", fmt.Sprint(id))
	}

	return s.GetStudentById(id)
}
