package handlers

import (
	"bytes"
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
	"unicode/utf8"

	"github.com/gin-gonic/gin"
)

const (
	AddUser = iota
	GetUser
	UpdateUser
	DeleteUser
)

func callUpdateUser(t *testing.T, handlers *Handlers, email string, input UpdateUserRequest) *httptest.ResponseRecorder {
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
	c.Params = gin.Params{gin.Param{
		Key:   "user_email",
		Value: email,
	}}
	handlers.UpdateUser(c)

	return r
}

func callDeleteUser(t *testing.T, handlers *Handlers, email string) *httptest.ResponseRecorder {
	t.Helper()

	r := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(r)
	c.Params = gin.Params{gin.Param{
		Key:   "user_email",
		Value: email,
	}}
	handlers.DeleteUser(c)

	return r
}

type State struct {
	InDB    bool
	Updated bool
}

func (s State) Log(t *testing.T) {
	t.Log("State, InDB:", s.InDB, "Updated:", s.Updated)
}

func FuzzSequenceHandlers(f *testing.F) {
	db := initDb(f)

	handlers := NewHandlers(db)

	// Add an example that adds a user, checks that the insert worked, and then deletes
	f.Add([]byte{0, 1, 3}, "john@smith.com", "John Smith", "+1 234 567 8901", "+9 876 543 2109")

	f.Fuzz(func(t *testing.T, operations []byte, email, name, number1, number2 string) {
		if len(email) == 0 {
			return
		}
		if !utf8.ValidString(email) || !utf8.ValidString(name) || !utf8.ValidString(number1) || !utf8.ValidString(number2) {
			return
		}
		defer func() {
			// We don't check the result, we just want to make sure the user is gone
			callDeleteUser(t, handlers, email)
		}()
		state := State{}

		// Limit the number of API calls we make in one go to 10
		if len(operations) > 10 {
			operations = operations[:10]
		}

		t.Log("Running", len(operations), "operations")
		for i, operation := range operations {
			t.Log("Operation:", i+1)
			switch operation % 4 {
			case AddUser:
				t.Log("Adding")
				input := AddUserRequest{
					User{
						Email:       email,
						Name:        name,
						PhoneNumber: number1,
					},
				}

				// Call AddUser with this data
				r := callAddUser(t, handlers, input)
				if r.Result().StatusCode != 200 {
					t.Log("Add failure")
					// This is ok, we want non-200 status codes for invalid data
					// We just shouldn't continue with the rest of the test
					return
				}
				t.Log("Add successful")
				state.InDB = true
				state.Updated = false
				state.Log(t)
			case GetUser:
				t.Log("Getting")
				// Here, the user has been inserted into the db
				// Call GetUser with the same email
				r := callGetUser(t, handlers, email)
				if state.InDB {
					if r.Result().StatusCode != 200 {
						t.Fatal("Expected to get user back, got", r.Result().StatusCode)
					}

					// Unmarshal the json response and check if the data is the same
					result := &GetUserResponse{}
					err := json.Unmarshal(r.Body.Bytes(), result)
					if err != nil {
						t.Fatal("Expected proper json response:", err)
					}
					expectedNumber := number1
					if state.Updated {
						// We expect the row to be updated
						expectedNumber = number2
					}
					if result.User.PhoneNumber != expectedNumber {
						t.Fatal("Expected number", expectedNumber, "got", result.User.PhoneNumber)
					}
				} else {
					// We don't expect the user to be in the db
					if r.Result().StatusCode != 404 {
						t.Fatal("Expected to get 404, got", r.Result().StatusCode)
					}
				}
				t.Log("Get successful")
				state.Log(t)
			case UpdateUser:
				t.Log("Updating")
				// Edit the user to insert a new phone number
				r := callUpdateUser(t, handlers, email, UpdateUserRequest{
					PhoneNumber: number2,
				})

				if r.Result().StatusCode != 200 {
					if !state.InDB {
						// Error code, but there's nothing to update in the db. This is good and expected
						t.Log("Attempted to update nonexistent row, returned non-200 status code")
					} else {
						bodyBs, err := ioutil.ReadAll(r.Result().Body)
						t.Fatal("Update should have succeeded, but failed", string(bodyBs), err)
					}
					state.InDB = false
					state.Updated = false
				} else {
					t.Log("Update successful")
					state.InDB = true
					state.Updated = true
				}
				state.Log(t)
			case DeleteUser:
				t.Log("Deleting")
				// Delete the user from the DB
				r := callDeleteUser(t, handlers, email)
				if r.Result().StatusCode == 200 {
					t.Log("Delete successful")
				} else {
					if !state.InDB {
						// The error was returned correctly. This is good and expected
						t.Log("Attempted to delete nonexistent row, returned non-200 status code")
					} else {
						t.Fatal("Delete should have succeeded, but failed")
					}
				}
				state.InDB = false
				state.Updated = false
				state.Log(t)
			}
		}
	})
}
