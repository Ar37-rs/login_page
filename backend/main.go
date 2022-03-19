package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"regexp"
	"time"

	rice "github.com/GeertJohan/go.rice"
	"github.com/go-playground/validator"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	_ "github.com/mattn/go-sqlite3"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

// Change these empty client_id and client_secret strings according to your GA ClientID and ClientSecret,
// you have to add the url: http://localhost:1323/auth/oauth2 as one of your Authorised redirect URIs on your GA credential config.
// for more info visit: https://console.cloud.google.com/apis/credentials
const client_id = ""
const client_secret = ""

const authorized = "Authorized"
const email_taken = "EmailTaken"
const invalid_password = "InvalidPassword"
const logged = "Logged"
const unauthorized = "Unauthorized"
const unregistered = "Unregistered"
const signup_accepted = "SignupAccepted"
const internal_error = "InternalErr"
const oauth2_url_state = "test123"

var Oauth2Config *oauth2.Config

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
	cookie.Path = "/"
	cookie.Value = ""
	cookie.Expires = time.Unix(0, 0)
	c.SetCookie(cookie)
}

func setCookie(cookie *http.Cookie, user User, c echo.Context) {
	cookie.Name = "username_logged"
	cookie.Path = "/"
	cookie.Value = user.Email
	cookie.Expires = time.Now().Add(24 * time.Hour)
	cookie.SameSite = http.SameSiteDefaultMode
	c.SetCookie(cookie)
}

func addUserOnline(user User) {
	db_im, _ := OpenDbIM()
	CreateTableIM(db_im)
	InsertDBIM(db_im, LoggedInfo{Email: user.Email, Status: "logged"})
	SelectAllDBIM(db_im)
}

func login_with_google(c echo.Context) error {
	url := Oauth2Config.AuthCodeURL(oauth2_url_state)

	// redirect to GA
	return c.Redirect(http.StatusTemporaryRedirect, url)
}

func GAuserInfo(state string, code string) (map[string]string, error) {
	if state != oauth2_url_state {
		return nil, fmt.Errorf("oauth2 state didn't match")
	}

	// exchange token
	token, err := Oauth2Config.Exchange(context.TODO(), code)
	if err != nil {
		return nil, err
	}

	res, err := http.Get("https://www.googleapis.com/oauth2/v2/userinfo?access_token=" + token.AccessToken)
	if err != nil {
		return nil, err
	}
	contents, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}
	data := make(map[string]string)
	json.Unmarshal([]byte(contents), &data)
	return data, nil
}

// handle oauth2 redirect
func handleRedirect(c echo.Context) error {
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
			return c.Redirect(http.StatusTemporaryRedirect, "/profile_view")
		}
	}
	data, err := GAuserInfo(c.FormValue("state"), c.FormValue("code"))
	if err != nil {
		return c.String(http.StatusUnauthorized, unauthorized)
	}

	// for temporary password: data["id"] + data["name"]
	pass := data["id"] + data["family_name"]
	user := User{Name: data["name"], Email: data["email"], Password: pass}
	db, err := OpenDb()
	if err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}
	accepted, _ := signupValidator(db, user)
	if accepted {
		if err := InsertDB(db, user); err != nil {
			return c.String(http.StatusInternalServerError, internal_error)
		} else {
			println("User inserted into DB")
		}
	}
	cookie = new(http.Cookie)
	setCookie(cookie, user, c)
	addUserOnline(user)
	return c.Redirect(http.StatusTemporaryRedirect, "/profile_view")
}

// to make SPA refresh work with go backend
func serveSPA(e *echo.Echo) {
	box, err := rice.FindBox("/build")
	if err != nil {
		log.Fatal("Unable to load SPA assets")
	}

	file, err := box.String("index.html")

	if err != nil {
		log.Fatalf("could not open file: %s\n", err)
	}

	assetHandler := http.FileServer(box.HTTPBox())
	e.GET("*", echo.WrapHandler(assetHandler), func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			matchAssets := regexp.MustCompile(`(.*)\.(js|png|json|ico|txt|css|svg)$`)
			if matchAssets.MatchString(c.Request().URL.RequestURI()) == true {
				return next(c)
			}
			return c.HTML(http.StatusOK, file)
		}
	})
}

func main() {
	e := echo.New()
	serveSPA(e)
	cors := middleware.CORSConfig{}
	cors.AllowCredentials = true
	e.Use(middleware.CORSWithConfig(cors))
	e.Validator = &CustomValidator{validator: validator.New()}

	// oauth2 config
	Oauth2Config = &oauth2.Config{
		RedirectURL:  "http://localhost:1323/auth/oauth2",
		ClientID:     client_id,
		ClientSecret: client_secret,
		Scopes:       []string{"https://www.googleapis.com/auth/userinfo.profile", "https://www.googleapis.com/auth/userinfo.email"},
		Endpoint:     google.Endpoint,
	}

	// login using google oauth2
	e.GET("/login_with_google", login_with_google)

	// oauth2: temporary redirect
	e.GET("/auth/oauth2", handleRedirect)

	// login manually, will be redirected to secret page if authorized (/profile_view).
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
			} else {
				return c.JSON(http.StatusOK, OutMessage{Message: authorized})
			}
		}
		user := new(User)
		if err := c.Bind(user); err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, err.Error())
		}
		if err := c.Validate(user); err != nil {
			return err
		}
		db, err := OpenDb()
		if err != nil {
			return c.String(http.StatusInternalServerError, err.Error())
		}
		defer db.Close()
		_, msg := loginValidator(db, *user)
		if msg == logged {
			cookie := new(http.Cookie)
			setCookie(cookie, *user, c)
			addUserOnline(*user)
			return c.JSON(http.StatusOK, OutMessage{Message: msg})
		}
		return c.JSON(http.StatusOK, OutMessage{Message: msg})

	})

	// signup page, will be redirected to secret page if authorized (/profile_view).
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
				return c.String(http.StatusInternalServerError, internal_error)
			} else {
				println("User inserted into DB")
			}
			cookie := new(http.Cookie)
			setCookie(cookie, *user, c)
			addUserOnline(*user)
			return c.JSON(http.StatusOK, OutMessage{Message: msg})
		}
		return c.JSON(http.StatusOK, OutMessage{Message: msg})
	})

	// logout API
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
		defer stmt.Close()
		var email string
		err = stmt.QueryRow(cookie.Value).Scan(&email)
		resetCookie(cookie, c)
		if err == nil {
			DeleteUserIM(db_im, email)
		}
		return c.JSON(http.StatusOK, OutMessage{Message: "Logged_out"})
	})

	// check if there's users online
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

	// secret page API
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
	// begin
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	stmt_logged, err := db.Prepare("select email from logged where email = ?")
	if err != nil {
		return err
	}
	defer stmt_logged.Close()
	var email string
	stmt_logged_err := stmt_logged.QueryRow(logged_info.Email).Scan(&email)
	if stmt_logged_err != nil {
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
func loginValidator(db *sql.DB, user User) (bool, string) {
	// query
	stmt, err := db.Prepare("select * from users where email = ?")
	if err != nil {
		return false, internal_error
	}
	defer stmt.Close()
	var name string
	var email string
	var password string
	err = stmt.QueryRow(user.Email).Scan(&name, &email, &password)
	if err != nil {
		return false, unregistered
	}

	if email == user.Email && password != user.Password {
		return false, invalid_password
	} else if email == user.Email && password == user.Password {
		return true, logged
	} else {
		return false, unregistered
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
