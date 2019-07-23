package main

import (
	"crypto/md5"
	"database/sql"
	"fmt"
	"github.com/labstack/echo"
	_ "github.com/lib/pq"
	"log"
	"net/http"
	"time"
)

func main() {
	var err error
	var db *sql.DB
	db, err = sql.Open("postgres", "user=postgres password=postgres dbname=mydb sslmode=disable")
	if err != nil{
		log.Fatal(err)
	}

	if err = db.Ping(); err != nil {
		panic(err)

	}else {
		fmt.Println("DB Connected")
	}

	e:=echo.New()


	//e.Use(middleware.Logger())
	//e.Use(middleware.Recover())

	type User struct {
		Id     string `json:"id"`
		Name   string `json:"name"`
		Email  string `json:"email"`
		Password   string `json:"password"`
	}

	type Users struct {
		users []User `json:"users"`
	}


	e.POST("/register", func(c echo.Context) error {
		u := new(User)
		if err := c.Bind(u); err != nil {
			return err
		}
		if u.Name == "" || u.Password == ""|| u.Email== ""{
			return c.JSON(http.StatusBadRequest,"missing data")


		}else {
			pass := []byte(u.Password)
			hashed_pass := fmt.Sprintf("%x", md5.Sum(pass))


			sqlStatement := "INSERT INTO user_data (name, email, password)VALUES ($1, $2, $3)"
			res, err := db.Query(sqlStatement, u.Name, u.Email, hashed_pass)

			if err != nil {
				fmt.Println(err)
				return c.JSON(http.StatusExpectationFailed, err)
			} else {
				fmt.Println(res)
				sqlStatement := "SELECT * from user_data where email = $1"
				user_data, err := db.Query(sqlStatement, u.Email)

				if err != nil {
					fmt.Println(err)
				}
				result := Users{}

				for user_data.Next() {
					data := User{}
					err2 := user_data.Scan(&data.Id, &data.Name, &data.Email, &data.Password)
					// Exit if we get an error
					if err2 != nil {
						return err2
					}
					result.users = append(result.users, data)
				}
				return c.JSON(http.StatusCreated, result.users)
				return c.JSON(http.StatusCreated, res)
			}

		}
	})


	e.POST("/login", func(c echo.Context) error {
		u := new(User)


		fmt.Println(u.Email)
		fmt.Println(u.Password)
		if err := c.Bind(u); err != nil {
			return err
		}
		if  u.Password == ""|| u.Email== ""{
			return c.JSON(http.StatusBadRequest,"missing data")

		}else {
			sqlStatement := "SELECT * FROM user_data WHERE email =$1"
			res, err := db.Query(sqlStatement,u.Email)
			if res.Next() == false {
				return c.JSON(http.StatusConflict, "user does not exist")

			}


			if err != nil {
				fmt.Println(err)
				return c.JSON(http.StatusExpectationFailed, err)
			} else if err == nil {


				for res.Next() {
					data := User{}
					err2 := res.Scan(&data.Id, &data.Name, &data.Email, &data.Password)
					pass := []byte(u.Password)
					hashed_pass := fmt.Sprintf("%x", md5.Sum(pass))
					fmt.Println(data.Email)

					// Exit if we get an error
					if err2 != nil {
						return err2
					}  else if data.Password != (hashed_pass) {
						return c.JSON(http.StatusBadRequest, "incorrect password")

					} else {
						cookie := new(http.Cookie)
						cookie.Name = "username"
						cookie.Value = data.Name
						cookie.Expires = time.Now().Add(24 * time.Hour)
						//cookie.Secure = true
						c.SetCookie(cookie)
						return c.JSON(http.StatusCreated, data.Name+" successfully logged in")
						//return c.JSON(http.StatusCreated, res)
					}
				}
			}

		}
		return c.JSON(http.StatusOK, err)


		//return c.String(http.StatusOK, "ok")
	})

	e.GET("/", func(c echo.Context) error {

		cookie, err := c.Cookie("username")
		if err != nil {
			return c.JSON(http.StatusBadRequest, "login first")

		}


		return c.String(http.StatusOK, "welcome "+ cookie.Value)
	})

	e.GET("/logout", func(c echo.Context) error {

		cookie, err := c.Cookie("username")
		if err != nil {
			return err
		}
		cookie.Expires= time.Now()
		c.SetCookie(cookie)

		return c.String(http.StatusOK, cookie.Value + " successfully logged out")
	})


	e.Start(":8889")

}


