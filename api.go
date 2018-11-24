package main

// GOOS=windows GOARCH=386 go build -o api32.exe

// james.bond	james123!

import (
	"errors"
	"math/rand"
	"net/http"
	"regexp"
	"strconv"
	"time"

	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
)

func init() {
	rand.Seed(time.Now().UnixNano())

	users = []User{
		User{
			Login:    "james.bond",
			Password: "james123!",
		},
	}
	sessions = make(Sessions, 0)
	employees = make([]Employee, 0)
}

var letterRunes = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

func RandStringRunes(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return string(b)
}

func main() {
	e := echo.New()
	e.Pre(middleware.RemoveTrailingSlash())

	bearerAuthentication := middleware.KeyAuth(func(token string, c echo.Context) (bool, error) {
		return sessions.Exists(token), nil
	})

	e.GET("/ping", func(c echo.Context) error {
		return c.String(http.StatusOK, "pong")
	})

	e.GET("/status", getStatus)

	e.POST("/login", login)
	e.DELETE("/logout", logout, bearerAuthentication)

	// This API can be only accessed with a token
	sr := e.Group("/employees")
	sr.Use(bearerAuthentication)
	sr.GET("/:id", getEmployee)
	sr.GET("", getAllEmployees)
	sr.POST("", createEmployee)
	sr.PUT("/:id", updateEmployee)
	sr.DELETE("/:id", deleteEmployee)

	e.Logger.Fatal(e.Start(":1323"))
}

type Session struct {
	token string
}

type Sessions []Session

type User struct {
	Login    string `json:"login" form:"login" query:"login"`
	Password string `json:"password" form:"password" query:"password"`
}

type Employee struct {
	ID       uint   `json:"id"`
	Fullname string `json:"fullName"`
	Age      uint8  `json:"age"`
	Email    string `json:"email"`
}

type Employees []Employee

var (
	users     []User
	sessions  Sessions
	employees Employees
)

func (s *Sessions) Add(token string) error {
	*s = append(*s, Session{token: token})
	return nil
}

func (s *Sessions) Remove(token string) error {
	for i, session := range *s {
		if session.token == token {
			*s = append((*s)[:i], (*s)[i+1:]...)
			return nil
		}
	}

	return nil
}

func (s Sessions) Exists(token string) bool {
	for _, session := range s {
		if session.token == token {
			return true
		}
	}

	return false
}

type TokenResponse struct {
	Token string `json:"token"`
}

func login(c echo.Context) error {
	u := new(User)
	if err := c.Bind(u); err != nil {
		return c.String(http.StatusBadRequest, "...")
	}

	for _, user := range users {
		if user.Login == u.Login && user.Password == u.Password {
			token := RandStringRunes(55)
			sessions.Add(token)
			return c.JSON(http.StatusOK, TokenResponse{Token: token})
		}
	}

	return c.String(http.StatusUnauthorized, "401 - Unauthorized")

	// name := c.FormValue("name")
	// email := c.FormValue("email")
	// return c.String(http.StatusOK, "...") //"name:"+u.Name+", email:"+u.Email+" .."+c.Request().Header.Get("Authorization")
}

func logout(c echo.Context) error {
	token := c.Request().Header.Get("Authorization")
	re := regexp.MustCompile(`[Bb]earer\s+`)
	token = re.ReplaceAllString(token, "")

	sessions.Remove(token)
	return c.String(http.StatusOK, "You were logout successfully")
}

func getStatus(c echo.Context) error {
	type SessionsResponse struct {
		Health        string `json:"health"`
		SessionsCount int    `json:"sessionsCount"`
	}

	return c.JSON(http.StatusOK, SessionsResponse{Health: "OK", SessionsCount: len(sessions)})
}

func getAllEmployees(c echo.Context) error {
	return c.JSON(http.StatusOK, employees)
}

func createEmployee(c echo.Context) error {
	e := new(Employee)
	if err := c.Bind(e); err != nil {
		return c.String(http.StatusBadRequest, "...")
	}
	if len(employees) == 0 {
		e.ID = 0
	} else {
		e.ID = employees[len(employees)-1].ID + 1
	}
	employees = append(employees, *e)
	return c.String(http.StatusCreated, "OK")
}

func (e Employees) Get(id uint) (Employee, error) {
	for _, employee := range e {
		if employee.ID == id {
			return employee, nil
		}
	}

	return Employee{}, errors.New("Employee could not be found")
}

func (e *Employees) Remove(id uint) error {
	for i, employee := range *e {
		if employee.ID == id {
			*e = append((*e)[:i], (*e)[i+1:]...)
			return nil
		}
	}

	return errors.New("Employee could not be found")
}

func getEmployee(c echo.Context) error {
	id, _ := strconv.Atoi(c.Param("id"))
	e, err := employees.Get(uint(id))
	if err != nil {
		return c.String(http.StatusNotFound, "Employee could not be found")
	}

	return c.JSON(http.StatusOK, e)
}

func updateEmployee(c echo.Context) error {
	id, _ := strconv.Atoi(c.Param("id"))

	e, err := employees.Get(uint(id))
	if err != nil {
		return c.String(http.StatusNotFound, "Employee could not be found")
	}

	fullName := c.FormValue("fullName")
	age := c.FormValue("age")
	email := c.FormValue("email")

	if fullName != "" {
		e.Fullname = fullName
	}

	if age != "" {
		ageInt32, _ := strconv.Atoi(age)
		e.Age = uint8(ageInt32)
	}

	if email != "" {
		e.Email = email
	}

	for i, employee := range employees {
		if employee.ID == e.ID {
			employees[i] = e
		}
	}

	return c.String(http.StatusCreated, "OK")
}

func deleteEmployee(c echo.Context) error {
	id, _ := strconv.Atoi(c.Param("id"))
	e, err := employees.Get(uint(id))
	if err != nil {
		return c.String(http.StatusNotFound, "Employee could not be found")
	}
	err = employees.Remove(e.ID)
	if err != nil {
		return c.String(http.StatusNotFound, "Employee could not be found")
	}
	return c.String(http.StatusOK, "OK")
}
