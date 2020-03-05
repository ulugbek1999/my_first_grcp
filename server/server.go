package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"time"

	"github.com/golang/protobuf/ptypes"
	_ "github.com/lib/pq"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	"github.com/jmoiron/sqlx"
	pb "github.com/ulugbek1999/my_first_grcp/pb"
)

var (
	// DB is an instance of sqlx.DB
	DB  *sqlx.DB
	err error
)

const (
	host     = "localhost"
	dbname   = "grpc_db"
	user     = "jeyran"
	password = "12345"
)

type student struct {
	pb.Student
	pb.UnimplementedStudentTextServer
}

type teacher struct {
	pb.Teacher
	pb.UnimplementedTeacherTextServer
}

func (s *student) Get(ctx context.Context, in *pb.Request) (std *pb.Student, err error) {
	st := pb.Student{}
	course := pb.Course{}
	st.Course = &course

	sm := time.Time{}

	stmt := `SELECT id, first_name, last_name, dob, course_id FROM student WHERE id = $1`
	err = DB.QueryRow(stmt, in.Id).
		Scan(&st.Id, &st.FirstName, &st.LastName, &sm, &st.Course.Id)
	if err != nil {
		log.Printf("Cannot get student: %s\n", err)
	}
	st.DoB, err = ptypes.TimestampProto(sm)
	if err != nil {
		log.Printf("Cannot get student's dob to timestamp: %s\n", err)
	}

	return &st, nil
}

func (s *student) GetAll(ctx context.Context, in *pb.Request) (stds *pb.Students, err error) {
	stds = &pb.Students{}
	students := []*pb.Student{}
	statement := `SELECT id, first_name, last_name, dob, course_id FROM student`
	rows, err := DB.Query(statement)
	if err != nil {
		log.Println("Cannot get students: ", err.Error())
	}
	st := &pb.Student{}
	cs := &pb.Course{}
	st.Course = cs
	sm := time.Time{}

	for rows.Next() {
		err = rows.Scan(
			&st.Id,
			&st.FirstName,
			&st.LastName,
			&sm,
			&st.Course.Id,
		)
		if err != nil {
			log.Panicln(err)
			return
		}
		st.DoB, err = ptypes.TimestampProto(sm)
		students = append(students, st)

	}
	defer rows.Close()

	stds.Students = students
	return
}

// Register method for student
func (s *student) Register(ctx context.Context, in *pb.Student) (std *pb.Response, err error) {
	stmt := `INSERT INTO student (first_name, last_name, dob, course_id)
			 VALUES ($1, $2, $3, $4) RETURNING id`

	tx, err := DB.Begin()
	if err != nil {
		tx.Rollback()
		log.Println("Cannot begin SQL transaction", err)
		std = &pb.Response{Message: "Cannot start transaction", Code: http.StatusBadRequest}
		return
	}

	err = tx.QueryRow(stmt, in.FirstName, in.LastName, time.Unix(in.DoB.GetSeconds(), 0), in.Course.Id).
		Scan(&in.Id)
	if err != nil {
		tx.Rollback()
		log.Println("Cannot insert student", err)
		std = &pb.Response{Message: "Cannot insert student", Code: http.StatusBadRequest}
		return
	}

	if int(in.Id) == 0 {
		std = &pb.Response{Message: "Something went wrong", Code: http.StatusBadRequest}
		tx.Rollback()
		return
	}

	std = &pb.Response{Message: "Successfully created", Code: http.StatusOK}

	tx.Commit()
	return
}

func (s *student) Edit(ctx context.Context, in *pb.Student) (std *pb.Response, err error) {
	stmt := `UPDATE student SET first_name = $2, last_name = $3, dob = $4, course_id = $5
			 WHERE id = $1`

	tx, err := DB.Begin()
	if err != nil {
		tx.Rollback()
		log.Println("Cannot begin SQL transaction", err)
		std = &pb.Response{Message: "Cannot begin SQL transaction", Code: http.StatusBadRequest}
		return
	}

	fmt.Println(in.Id)
	_, err = tx.Exec(stmt, in.Id, in.FirstName, in.LastName, time.Unix(in.DoB.GetSeconds(), 0), in.Course.Id)
	if err != nil {
		tx.Rollback()
		log.Println("Cannot update student information", err)
		std = &pb.Response{Message: "Cannot update student information", Code: http.StatusBadRequest}
		return
	}

	std = &pb.Response{Message: "Successfully updated", Code: http.StatusOK}
	tx.Commit()
	return
}

func (s *student) Remove(ctx context.Context, in *pb.Request) (std *pb.Response, err error) {
	stmt := `DELETE FROM student WHERE id = $1 `
	_, err = DB.Exec(stmt, in.Id)
	if err != nil {
		std = &pb.Response{Message: "Cannot delete", Code: http.StatusBadRequest}
	} else {
		std = &pb.Response{Message: "Successfully deleted", Code: http.StatusOK}
	}
	return
}

// Get method for teacher
func (t *teacher) Get(ctx context.Context, in *pb.Request) (std *pb.Teacher, err error) {
	tch := pb.Teacher{}

	dob := time.Time{}
	jd := time.Time{}

	stmt := `SELECT id, first_name, last_name, dob, joined_date FROM teacher WHERE id = $1`
	err = DB.QueryRow(stmt, in.Id).
		Scan(&tch.Id, &tch.FirstName, &tch.LastName, &dob, &jd)
	if err != nil {
		log.Printf("Cannot get teacher: %s\n", err)
	}

	tch.DoB, err = ptypes.TimestampProto(dob)
	if err != nil {
		log.Printf("Cannot convert teacher's dob to timestamp: %s\n", err)
	}

	tch.JoinedDate, err = ptypes.TimestampProto(jd)
	if err != nil {
		log.Printf("Cannot convert teacher's joined_date to timestamp: %s\n", err)
	}

	return &tch, nil
}

// GetAll
func (t *teacher) GetAll(ctx context.Context, in *pb.Request) (tchs *pb.Teachers, err error) {
	tchs = &pb.Teachers{}
	teachers := []*pb.Teacher{}
	statement := `SELECT id, first_name, last_name, dob, joined_date FROM teacher`
	rows, err := DB.Query(statement)
	if err != nil {
		log.Println("Cannot get students: ", err.Error())
	}
	tch := &pb.Teacher{}
	sm := time.Time{}
	jd := time.Time{}

	for rows.Next() {
		err = rows.Scan(
			&tch.Id,
			&tch.FirstName,
			&tch.LastName,
			&sm,
			&jd,
		)
		if err != nil {
			log.Panicln(err)
			return
		}
		tch.DoB, err = ptypes.TimestampProto(sm)
		tch.JoinedDate, err = ptypes.TimestampProto(jd)
		teachers = append(teachers, tch)

	}
	defer rows.Close()

	tchs.Teachers = teachers
	return
}

// Edit method for teacher
func (t *teacher) Edit(ctx context.Context, in *pb.Teacher) (std *pb.Response, err error) {
	stmt := `UPDATE teacher SET first_name = $2, last_name = $3, dob = $4, joined_date = $5
			 WHERE id = $1`

	tx, err := DB.Begin()
	if err != nil {
		tx.Rollback()
		log.Println("Cannot begin SQL transaction", err)
		std = &pb.Response{Message: "Cannot begin SQL transaction", Code: http.StatusBadRequest}
		return
	}

	fmt.Println(in.Id)
	_, err = tx.Exec(stmt, in.Id, in.FirstName, in.LastName, time.Unix(in.DoB.GetSeconds(), 0), time.Unix(in.JoinedDate.GetSeconds(), 0))
	if err != nil {
		tx.Rollback()
		log.Println("Cannot update teacher information", err)
		std = &pb.Response{Message: "Cannot update teacher information", Code: http.StatusBadRequest}
		return
	}

	std = &pb.Response{Message: "Successfully updated", Code: http.StatusOK}
	tx.Commit()
	return
}

// Register method for teacher
func (t *teacher) Register(ctx context.Context, in *pb.Teacher) (std *pb.Response, err error) {
	stmt := `INSERT INTO teacher (first_name, last_name, dob, joined_date)
			 VALUES ($1, $2, $3, $4) RETURNING id`

	tx, err := DB.Begin()
	if err != nil {
		tx.Rollback()
		log.Println("Cannot begin SQL transaction", err)
		std = &pb.Response{Message: "Cannot start transaction", Code: http.StatusBadRequest}
		return
	}

	err = tx.QueryRow(stmt, in.FirstName, in.LastName, time.Unix(in.DoB.GetSeconds(), 0), time.Unix(in.JoinedDate.GetSeconds(), 0)).
		Scan(&in.Id)
	if err != nil {
		tx.Rollback()
		log.Println("Cannot insert student", err)
		std = &pb.Response{Message: "Cannot insert student", Code: http.StatusBadRequest}
		return
	}

	if int(in.Id) == 0 {
		std = &pb.Response{Message: "Something went wrong", Code: http.StatusBadRequest}
		tx.Rollback()
		return
	}

	std = &pb.Response{Message: "Successfully created", Code: http.StatusOK}

	tx.Commit()
	return
}

// Remove method for teacher
func (t *teacher) Remove(ctx context.Context, in *pb.Request) (std *pb.Response, err error) {
	stmt := `DELETE FROM teacher WHERE id = $1 `
	_, err = DB.Exec(stmt, in.Id)
	if err != nil {
		std = &pb.Response{Message: "Cannot delete", Code: http.StatusBadRequest}
	} else {
		std = &pb.Response{Message: "Successfully deleted", Code: http.StatusOK}

	}
	return
}

func init() {
	log.Println("Connecting to the database...")
	connectionString := fmt.Sprintf("host=%s port=5432 dbname=%s user=%s password=%s sslmode=disable",
		host,
		dbname,
		user,
		password,
	)
	for i := 0; i < 10; i++ {
		DB, err = sqlx.Connect("postgres", connectionString)
		if err == nil {
			break
		}

		time.Sleep(1 * time.Second)
	}

	if err != nil {
		log.Fatalln(err)
	}

	if err = DB.Ping(); err != nil {
		log.Fatalln(err)
	}
	log.Println("Successfully connected!")
}

func main() {
	lis, err := net.Listen("tcp", ":4555")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	s := grpc.NewServer()
	pb.RegisterStudentTextServer(s, &student{})
	pb.RegisterTeacherTextServer(s, &teacher{})

	// Register reflection service on gRPC server.
	reflection.Register(s)
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
