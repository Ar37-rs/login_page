package main

import (
	"database/sql"
	"fmt"
	"net/http"
	"time"

	"github.com/go-playground/validator"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	_ "github.com/mattn/go-sqlite3"
)

const authorized = "Authorized"
const email_taken = "EmailTaken"
const invalid_password = "InvalidPassword"
const logged = "Logged"
const unauthorized = "Unauthorized"
const unregistered = "Unregistered"
const signup_accepted = "SignupAccepted"
const internal_error = "InternalErr"

type (
	User struct {
		Name     string `json:"name"`
		Email    string `json:"email" validate:"required,email"`
		Password string `json:"password" validate:"required"`
	}

	LoggedInfo struct {
		Email  string `json:"email"`
		Status string `json:"status"`
	}

	ProfileInfo struct {
		Name  string `json:"name"`
		Email string `json:"email"`
	}

	OutMessage struct {
		Message string `json:"message"`
	}

	CustomValidator struct {
		validator *validator.Validate
	}
)

func (cv *CustomValidator) Validate(i interface{}) error {
	if err := cv.validator.Struct(i); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return nil
}

func resetCookie(cookie *http.Cookie, c echo.Context) {
	cookie.Name = "username_logged"
	cookie.Value = ""
	cookie.Expires = time.Unix(0, 0)
	c.SetCookie(cookie)
}
func main() {
	e := echo.New()
	e.Static("/", "build")
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowCredentials: true,
		AllowOrigins: []string{
			"http://localhost:3000",
			"http://localhost:1323",
		},
		AllowHeaders: []string{echo.HeaderOrigin, echo.HeaderContentType, echo.HeaderAccept},
	}))
	e.Validator = &CustomValidator{validator: validator.New()}

	// masuk (login), akan redirected ke secret page jika authorized (difrontend).
	e.POST("/login_api", func(c echo.Context) (err error) {
		cookie, err := c.Cookie("username_logged")
		if err == nil {
			count := 0
			db_im, _ := OpenDbIM()
			rows, _ := db_im.Query("select email, status from logged")
			for rows.Next() {
				count += 1
			}
			rows.Close()
			if count <= 0 {
				resetCookie(cookie, c)
			}
			return c.JSON(http.StatusOK, OutMessage{Message: authorized})
		}
		user := new(User)
		if err := c.Bind(user); err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, err.Error())
		}
		if err := c.Validate(user); err != nil {
			return err
		}
		db, _ := OpenDb()
		defer db.Close()
		_, msg, _ := loginValidator(db, *user)
		if msg == logged {
			cookie := new(http.Cookie)
			cookie.Name = "username_logged"
			cookie.Value = user.Email
			cookie.Expires = time.Now().Add(24 * time.Hour)
			cookie.SameSite = http.SameSiteDefaultMode
			c.SetCookie(cookie)
			db_im, _ := OpenDbIM()
			CreateTableIM(db_im)
			InsertDBIM(db_im, LoggedInfo{Email: user.Email, Status: "logged"})
			SelectAllDBIM(db_im)
			return c.JSON(http.StatusOK, OutMessage{Message: msg})
		} else {
			return c.JSON(http.StatusOK, OutMessage{Message: msg})
		}
	})

	// api logout
	e.GET("/logout_api", func(c echo.Context) (err error) {
		cookie, err := c.Cookie("username_logged")
		if err != nil {
			return c.String(http.StatusOK, "empty result")
		}
		db_im, err := OpenDbIM()
		if err != nil {
			return c.String(http.StatusInternalServerError, err.Error())
		}
		stmt, err := db_im.Prepare("select email from logged where email = ?")
		if err != nil {
			return c.String(http.StatusInternalServerError, err.Error())
		}
		var email string
		err = stmt.QueryRow(cookie.Value).Scan(&email)
		cookie.Name = "username_logged"
		cookie.Value = ""
		cookie.Expires = time.Unix(0, 0)
		c.SetCookie(cookie)
		if err == nil {
			DeleteUserIM(db_im, email)
			stmt.Close()
		}
		return c.JSON(http.StatusOK, "logout")
	})

	// cek jika ada users yang masuk
	e.GET("/users_logged", func(c echo.Context) (err error) {
		db, err := OpenDbIM()
		if err != nil {
			return c.String(http.StatusInternalServerError, err.Error())
		}
		rows, err := db.Query("select email, status from logged")
		if err != nil {
			return c.String(http.StatusInternalServerError, err.Error())
		}
		defer rows.Close()
		users_logged := make(map[string]string)
		for rows.Next() {
			var email string
			var status string
			_ = rows.Scan(&email, &status)
			users_logged[email] = status
		}
		return c.JSON(http.StatusOK, users_logged)
	})

	// halaman secret page.
	e.GET("/profile", func(c echo.Context) (err error) {
		cookie, err := c.Cookie("username_logged")
		if err != nil {
			return c.JSON(http.StatusUnauthorized, OutMessage{Message: unauthorized})
		}

		count := 0
		db_im, _ := OpenDbIM()
		rows, _ := db_im.Query("select email, status from logged")
		for rows.Next() {
			count += 1
		}
		rows.Close()
		if count <= 0 {
			resetCookie(cookie, c)
		}

		db, err := OpenDb()
		if err != nil {
			return c.String(http.StatusInternalServerError, err.Error())
		}
		stmt, err := db.Prepare("select name, email from users where email = ?")
		if err != nil {
			return c.String(http.StatusInternalServerError, err.Error())
		}
		defer db.Close()
		var name string
		var email string
		err = stmt.QueryRow(cookie.Value).Scan(&name, &email)
		stmt.Close()
		if err != nil {
			return c.JSON(http.StatusUnauthorized, OutMessage{Message: unauthorized})
		}
		return c.JSON(http.StatusOK, ProfileInfo{name, email})
	})

	// halaman pendaftar, akan redirected ke secret page jika authorized (difrontend)
	e.POST("/signup_api", func(c echo.Context) (err error) {
		cookie, err := c.Cookie("username_logged")
		if err == nil {
			count := 0
			db_im, _ := OpenDbIM()
			rows, _ := db_im.Query("select email, status from logged")
			for rows.Next() {
				count += 1
			}
			rows.Close()
			if count <= 0 {
				resetCookie(cookie, c)
			} else {
				return c.JSON(http.StatusOK, OutMessage{Message: authorized})
			}
		}

		user := new(User)
		if err := c.Bind(user); err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, err.Error())
		}
		if err := c.Validate(user); err != nil {
			return c.JSON(http.StatusBadRequest, err.Error())
		}
		db, err := OpenDb()
		if err != nil {
			return c.String(http.StatusInternalServerError, err.Error())
		}
		defer db.Close()
		accepted, msg := signupValidator(db, *user)
		if accepted {
			if err := InsertDB(db, *user); err != nil {
				return c.JSON(http.StatusInternalServerError, internal_error)
			} else {
				println("User inserted into DB")
			}
			cookie := new(http.Cookie)
			cookie.Name = "username_logged"
			cookie.Value = user.Email
			cookie.Expires = time.Now().Add(24 * time.Hour)
			cookie.SameSite = http.SameSiteDefaultMode
			c.SetCookie(cookie)
			db_im, _ := OpenDbIM()
			CreateTableIM(db_im)
			InsertDBIM(db_im, LoggedInfo{Email: user.Email, Status: "logged"})
			SelectAllDBIM(db_im)
			return c.JSON(http.StatusOK, OutMessage{Message: msg})
		}
		return c.JSON(http.StatusOK, OutMessage{Message: msg})
	})

	e.Logger.Fatal(e.Start(":1323"))
}

func OpenDb() (*sql.DB, error) {
	db, err := sql.Open("sqlite3", "./database.sqlite3")
	if err != nil {
		return nil, err
	}
	err2 := CreateTable(db)
	if err2 != nil {
		return nil, err2
	}
	return db, err

}

// Open DB in memory
func OpenDbIM() (*sql.DB, error) {
	db, err := sql.Open("sqlite3", "file::memory:?cache=shared")
	if err != nil {
		return nil, err
	}
	err2 := CreateTableIM(db)
	if err2 != nil {
		return nil, err2
	}
	return db, err

}

func CreateTable(db *sql.DB) error {
	sqlStmt := `
	CREATE TABLE IF NOT EXISTS users(name VARCHAR(100), email VARCHAR(100), password VARCHAR(100));
	`
	_, err := db.Exec(sqlStmt)
	return err
}

// Create table in memory
func CreateTableIM(db *sql.DB) error {
	sqlStmt := `
	CREATE TABLE IF NOT EXISTS logged(email VARCHAR(100), status VARCHAR(100));
	`
	_, err := db.Exec(sqlStmt)
	return err
}

// Insert user into databse in memory
func InsertDB(db *sql.DB, user User) error {
	// beging
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	stmt, err := tx.Prepare("insert into users(name, email, password) values(?, ?, ?)")
	if err != nil {
		return err
	}
	defer stmt.Close()
	_, err = stmt.Exec(user.Name, user.Email, user.Password)
	if err != nil {
		return err
	}
	return tx.Commit()
}

// Insert user into databse in memory
func InsertDBIM(db *sql.DB, logged_info LoggedInfo) error {
	// beging
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	stmt_logged, err := db.Prepare("select email from logged where email = ?")
	if err != nil {
		return err
	}
	var email string
	stmt_logged_err := stmt_logged.QueryRow(logged_info.Email).Scan(&email)
	if stmt_logged_err != nil {
		stmt_logged.Close()
		stmt, err := tx.Prepare("insert into logged(email, status) values(?, ?)")
		if err != nil {
			return err
		}
		defer stmt.Close()
		_, err = stmt.Exec(logged_info.Email, logged_info.Status)
		if err != nil {
			return err
		}
		return tx.Commit()
	}
	return nil
}

// Query all database in memory
func SelectAllDBIM(db *sql.DB) error {
	rows, err := db.Query("select email, status from logged")
	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {

		var email string
		var status string
		err = rows.Scan(&email, &status)
		if err != nil {
			return err
		}
		fmt.Println(email, status)
	}
	return rows.Err()
}

// Query all database
func SelectAllDB(db *sql.DB, user User) error {
	rows, err := db.Query("select name, email, password from users")
	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {

		var name string
		var email string
		var password string
		err = rows.Scan(&name, &email, &password)
		if err != nil {
			return err
		}
		fmt.Println(name, email, password)
	}
	return rows.Err()
}

// Validate login
func loginValidator(db *sql.DB, user User) (bool, string, string) {
	// // query
	stmt, err := db.Prepare("select * from users where email = ?")
	if err != nil {
		return false, internal_error, ""
	}
	defer stmt.Close()
	var name string
	var email string
	var password string
	err = stmt.QueryRow(user.Email).Scan(&name, &email, &password)
	if err != nil {
		return false, unregistered, ""
	}

	if email == user.Email && password != user.Password {
		return false, invalid_password, ""
	} else if email == user.Email && password == user.Password {
		return true, logged, name
	} else {
		return false, unregistered, ""
	}
}

// Validate signup
func signupValidator(db *sql.DB, user User) (bool, string) {

	// query
	stmt, err := db.Prepare("select email, password from users where email = ?")
	if err != nil {
		return false, internal_error
	}
	defer stmt.Close()
	var email string
	var password string
	err = stmt.QueryRow(user.Email).Scan(&email, &password)
	if err != nil {
		return true, signup_accepted
	}

	if email != user.Email && password != user.Password {
		return true, signup_accepted
	} else if email == user.Email {
		return false, email_taken
	} else {
		return false, internal_error
	}
}

// Delete user from DB
func DeleteUser(db *sql.DB, user User) error {
	// delete
	stmt, err := db.Prepare("delete from users where email = ?")
	if err != nil {
		return err
	}
	defer stmt.Close()
	var email string
	err = stmt.QueryRow(user.Email).Scan(&email)
	if err != nil {
		return err
	}
	return nil
}

// Delete user from DB in Memory
func DeleteUserIM(db *sql.DB, db_in_email string) error {
	// delete
	stmt, err := db.Prepare("delete from logged where email = ?")
	if err != nil {
		return err
	}
	defer stmt.Close()
	var email string
	err = stmt.QueryRow(db_in_email).Scan(&email)
	if err != nil {
		return err
	}
	return nil
}