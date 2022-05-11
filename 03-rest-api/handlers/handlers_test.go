package handlers

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"unicode/utf8"

	_ "github.com/mattn/go-sqlite3"

	"github.com/gin-gonic/gin"
)

func initDb(t testing.TB) *sql.DB {
	t.Helper()

	// Set up an in-memory DB for testing
	db, err := sql.Open("sqlite3", "file::memory:?cache=shared")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		db.Close()
	})
	err = MigrateDB(db)
	if err != nil {
		t.Fatal(err)
	}
	return db
}

func callAddUser(t *testing.T, handlers *Handlers, input AddUserRequest) *httptest.ResponseRecorder {
	t.Helper()

	jsonBody, err := json.Marshal(input)
	if err != nil {
		t.Fatal(err)
	}

	r := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(r)
	req := &http.Request{
		Body: io.NopCloser(bytes.NewBuffer(jsonBody)),
	}
	c.Request = req
	handlers.AddUser(c)

	return r
}

func callGetUser(t *testing.T, handlers *Handlers, email string) *httptest.ResponseRecorder {
	t.Helper()

	r := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(r)
	c.Params = gin.Params{gin.Param{
		Key:   "user_email",
		Value: email,
	}}
	handlers.GetUser(c)

	return r
}

func TestInsertUser(t *testing.T) {
	db := initDb(t)
	handlers := NewHandlers(db)

	// Set up input data
	input := AddUserRequest{
		User{
			Name:        "John Smith",
			Email:       "john@smith.com",
			PhoneNumber: "+1 650 750 8505",
		},
	}

	// Call AddUser with this data
	r := callAddUser(t, handlers, input)
	if r.Result().StatusCode != 200 {
		t.Fatal("Expected insert user to work")
	}

	// Call GetUser with the same email
	r = callGetUser(t, handlers, input.Email)
	if r.Result().StatusCode != 200 {
		t.Fatal("Expected to get user back")
	}

	// Unmarshal the json response and check if the data is the same
	result := &GetUserResponse{}
	err := json.Unmarshal(r.Body.Bytes(), result)
	if err != nil {
		t.Fatal("Expected proper json response:", err)
	}
	if result.User.Name != input.Name {
		t.Fatal("Expected name", input.Name, "got", result.User.Name)
	}
}

func FuzzInsertUser(f *testing.F) {
	// Initialize the DB outside of the fuzz loop to save time
	db := initDb(f)
	// Set up our handlers with the in-memory test DB
	handlers := NewHandlers(db)

	// Add the example inputs from the unit test
	f.Add("john@smith.com", "John Smith", "+1 234 567 8901")

	f.Fuzz(func(t *testing.T, email, name, phoneNumber string) {
		// We validate that all of the inputs are correct UTF-8 strings. If we don't do this, errors
		// in the communication with SQLite manifest. (This is likely a bug that the fuzzing inadvertently
		// turned up, and is worth looking into, but outside the scope of this tutorial)
		if !utf8.ValidString(email) || !utf8.ValidString(name) || !utf8.ValidString(phoneNumber) {
			return
		}
		defer func() {
			// We don't check the result, we just want to make sure the user is gone
			// after each iteration of the test, so we start off with an empty DB
			callDeleteUser(t, handlers, email)
		}()

		// Set up input data
		input := AddUserRequest{
			User{
				Email:       email,
				Name:        name,
				PhoneNumber: phoneNumber,
			},
		}

		// Call AddUser with this data
		r := callAddUser(t, handlers, input)
		if r.Result().StatusCode != 200 {
			// This is ok, we want non-200 status codes for invalid data
			// We just shouldn't continue with the rest of the test
			return
		}

		// Here, the user has been inserted into the db
		// Call GetUser with the same email
		r = callGetUser(t, handlers, input.Email)
		if r.Result().StatusCode != 200 {
			t.Fatal("Expected to get user back, got", r.Result().StatusCode)
		}

		// Unmarshal the json response and check if the data is the same
		result := &GetUserResponse{}
		err := json.Unmarshal(r.Body.Bytes(), result)
		if err != nil {
			t.Fatal("Expected proper json response:", err)
		}
		if result.User.Name != input.Name {
			t.Log("Expected:", []byte(input.Name), "got", []byte(result.User.Name))
			t.Fatal("Expected name", input.Name, "got", result.User.Name)
		}
	})
}
