package main

import (
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/golang/protobuf/ptypes"
	pb "github.com/ulugbek1999/my_first_grcp/pb"

	"github.com/gin-gonic/gin"
	"google.golang.org/grpc"
)

var (
	studentCLT pb.StudentTextClient
	teacherCLT pb.TeacherTextClient
)

func main() {
	// Set up a connection to the server.
	conn, err := grpc.Dial("localhost:4555", grpc.WithInsecure())
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()
	studentCLT = pb.NewStudentTextClient(conn)
	teacherCLT = pb.NewTeacherTextClient(conn)

	// Set up a http server.
	r := gin.Default()

	// Student api endpoints
	r.GET("/student/:id", getStudent)
	r.GET("/students/all", getAllStudents)
	r.POST("/student/register", registerStudent)
	r.PUT("/student/edit/:id", updateStudent)
	r.DELETE("/student/delete/:id", deleteStudent)

	// Teacher api endpoints
	r.GET("/teacher/:id", getTeacher)
	r.GET("/teachers/all", getAllTeachers)
	r.POST("/teacher/register", registerTeacher)
	r.PUT("/teacher/edit/:id", updateTeacher)
	r.DELETE("/teacher/delete/:id", deleteTeacher)

	// Run http server
	if err := r.Run(":8053"); err != nil {
		log.Fatalf("could not run server: %v", err)
	}
}

// Get student with specified id
func getStudent(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	// Contact the server and print out its response.
	req := &pb.Request{Id: int32(id)}
	res, err := studentCLT.Get(c, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, res)
}

// Get all students
func getAllStudents(c *gin.Context) {
	req := &pb.Request{}
	res, err := studentCLT.GetAll(c, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, res)
}

// Register student
func registerStudent(c *gin.Context) {
	fn := c.PostForm("first_name")
	ln := c.PostForm("last_name")
	dob, err := time.Parse(time.RFC3339, c.PostForm("dob"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
	}
	dobTS, err := ptypes.TimestampProto(dob)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
	}

	crID, err := strconv.Atoi(c.PostForm("course_id"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
	}
	cr := &pb.Course{Id: int32(crID)}

	req := &pb.Student{FirstName: fn, LastName: ln, DoB: dobTS, Course: cr}
	res, err := studentCLT.Register(c, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, res)
}

// Update student
func updateStudent(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	fn := c.PostForm("first_name")
	ln := c.PostForm("last_name")
	dob, err := time.Parse(time.RFC3339, c.PostForm("dob"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
	}
	dobTS, err := ptypes.TimestampProto(dob)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
	}

	crID, err := strconv.Atoi(c.PostForm("course_id"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
	}
	cr := &pb.Course{Id: int32(crID)}
	req := &pb.Student{Id: int32(id), FirstName: fn, LastName: ln, DoB: dobTS, Course: cr}
	res, err := studentCLT.Edit(c, req)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, res)

}

// Delete student
func deleteStudent(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	req := &pb.Request{Id: int32(id)}
	res, err := studentCLT.Remove(c, req)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, res)
}

// Get teacher with specified id
func getTeacher(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	// Contact the server and print out its response.
	req := &pb.Request{Id: int32(id)}
	res, err := teacherCLT.Get(c, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, res)
}

// Register student
func registerTeacher(c *gin.Context) {
	fn := c.PostForm("first_name")
	ln := c.PostForm("last_name")
	dob, err := time.Parse(time.RFC3339, c.PostForm("dob"))
	if err != nil {
		log.Println("Cannot parse to teacher's dob")
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
	}
	dobTS, err := ptypes.TimestampProto(dob)
	if err != nil {
		log.Println("Cannot convert to teacher's dob to timestamp")
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
	}

	jd := time.Now()
	jdTS, err := ptypes.TimestampProto(jd)

	req := &pb.Teacher{FirstName: fn, LastName: ln, DoB: dobTS, JoinedDate: jdTS}
	res, err := teacherCLT.Register(c, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, res)
}

// Update teacher
func updateTeacher(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	fn := c.PostForm("first_name")
	ln := c.PostForm("last_name")
	dob, err := time.Parse(time.RFC3339, c.PostForm("dob"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
	}
	dobTS, err := ptypes.TimestampProto(dob)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
	}

	jd, err := time.Parse(time.RFC3339, c.PostForm("joined_date"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
	}

	jdTS, err := ptypes.TimestampProto(jd)
	req := &pb.Teacher{Id: int32(id), FirstName: fn, LastName: ln, DoB: dobTS, JoinedDate: jdTS}
	res, err := teacherCLT.Edit(c, req)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, res)

}

// Delete teacher
func deleteTeacher(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	req := &pb.Request{Id: int32(id)}
	res, err := studentCLT.Remove(c, req)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, res)
}

// Get all teachers
func getAllTeachers(c *gin.Context) {
	req := &pb.Request{}
	res, err := teacherCLT.GetAll(c, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, res)
}
