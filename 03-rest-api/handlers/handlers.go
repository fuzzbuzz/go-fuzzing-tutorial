package handlers

import (
	"database/sql"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/mattn/go-sqlite3"
)

func NewHandlers(db *sql.DB) *Handlers {
	return &Handlers{
		db: db,
	}
}

type Handlers struct {
	db *sql.DB
}

type User struct {
	Email       string
	Name        string
	PhoneNumber string
}

type AddUserRequest struct {
	User
}

func (h *Handlers) AddUser(c *gin.Context) {
	request := &AddUserRequest{}
	err := c.BindJSON(request)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": fmt.Sprintf("Invalid request: %s", err),
		})
		return
	}

	_, err = h.db.Exec(fmt.Sprintf("insert into users(name, email, phone_number) values ('%s', '%s', '%s');",
		request.Name, request.Email, request.PhoneNumber))
	if err != nil && err == sqlite3.ErrConstraintUnique {
		// return a 400 if the user has already inserted this email
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "email already added",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"user": request,
	})
}

type GetUserResponse struct {
	User
}

func (h *Handlers) GetUser(c *gin.Context) {
	wantEmail, ok := c.Params.Get("user_email")
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Param 'user_email' required",
		})
		return
	}

	result := h.db.QueryRow("select * from users where email = ?;", wantEmail)
	var email, name, phone_number string
	err := result.Scan(&email, &name, &phone_number)
	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "Email " + wantEmail + " not found",
			})
		} else {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": err.Error(),
			})
		}
		return
	}

	c.JSON(http.StatusOK, GetUserResponse{
		User: User{
			Email:       email,
			Name:        name,
			PhoneNumber: phone_number,
		},
	})
}

type UpdateUserRequest struct {
	PhoneNumber string
}

type UpdateUserResponse struct {
	User
}

func (h *Handlers) UpdateUser(c *gin.Context) {
	wantEmail, ok := c.Params.Get("user_email")
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Param 'user_email' required",
		})
		return
	}

	request := &UpdateUserRequest{}
	err := c.BindJSON(request)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": fmt.Sprintf("Invalid request: %s", err),
		})
		return
	}

	_, err = h.db.Exec("update users set phone_number = ? where email = ?;", request.PhoneNumber, wantEmail)
	if err != nil {
		if err != sql.ErrNoRows {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": err.Error(),
			})
		}
		return
	}

	c.JSON(http.StatusOK, UpdateUserResponse{
		User: User{
			Email:       wantEmail,
			PhoneNumber: request.PhoneNumber,
		},
	})
}

func (h *Handlers) DeleteUser(c *gin.Context) {
	fmt.Println(c.Params)
	wantEmail, ok := c.Params.Get("user_email")
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Param 'user_email' required",
		})
		return
	}

	_, err := h.db.Exec("delete from users where email = ?;", wantEmail)
	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "Email " + wantEmail + " not found",
			})
		} else {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": err.Error(),
			})
		}
		return
	}

	c.Status(http.StatusOK)
}
