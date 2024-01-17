package auth

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"
	"voxeti/backend/schema"

	"github.com/gorilla/sessions"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/integration/mtest"
)

func TestLogin(t *testing.T) {
	assert := assert.New(t)

	// mocking the echo context
	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/", nil)
	res := httptest.NewRecorder()
	context := e.NewContext(req, res)
	// initializing cookie store
	var store = sessions.NewCookieStore([]byte("test"))

	// Create DB Test Cases:
	testCases := []struct {
		name             string
		credentials      schema.Credentials
		prepMongoMock    func(mt *mtest.T)
		expectedResponse schema.LoginResponse
		expectedError    schema.ErrorResponse
		wantError        bool
	}{
		{
			name: "Success",
			credentials: schema.Credentials{
				Email:    "user1@example.com",
				Password: "password1",
			},
			prepMongoMock: func(mt *mtest.T) {
				user := schema.User{
					Email:          "user1@example.com",
					Password:       "$2a$10$yQMzszWR14B7a8WmQh4GT.gf4bf/x1ntXpX0kobFKIW8kOHQ2DOji",
					SocialProvider: "NONE",
				}

				userBSON, _ := bson.Marshal(user)
				var bsonD bson.D
				err := bson.Unmarshal(userBSON, &bsonD)
				if err != nil {
					assert.Fail("Failed to unmarshal bson data into document while prepping mock mongoDB. Method: 'Success'")
				}

				res := mtest.CreateCursorResponse(
					1,
					"data.users",
					mtest.FirstBatch,
					bsonD)
				end := mtest.CreateCursorResponse(
					0,
					"data.users",
					mtest.NextBatch)
				mt.AddMockResponses(res, end)
			},
			expectedResponse: schema.LoginResponse{
				User: schema.User{
					Email:    "user1@example.com",
					Password: "$2a$10$yQMzszWR14B7a8WmQh4GT.gf4bf/x1ntXpX0kobFKIW8kOHQ2DOji",
				},
			},
			wantError: false,
		},
		{
			name: "Invalid Password",
			credentials: schema.Credentials{
				Email:    "user1@example.com",
				Password: "password1",
			},
			prepMongoMock: func(mt *mtest.T) {
				user := schema.User{
					Email:          "user1@example.com",
					Password:       "someRandomPassword",
					SocialProvider: "NONE",
				}

				userBSON, _ := bson.Marshal(user)
				var bsonD bson.D
				err := bson.Unmarshal(userBSON, &bsonD)
				if err != nil {
					assert.Fail("Failed to unmarshal bson data into document while prepping mock mongoDB. Method: 'Invalid Password'")
				}

				res := mtest.CreateCursorResponse(
					1,
					"data.users",
					mtest.FirstBatch,
					bsonD)
				end := mtest.CreateCursorResponse(
					0,
					"data.users",
					mtest.NextBatch)
				mt.AddMockResponses(res, end)
			},
			expectedError: schema.ErrorResponse{
				Code:    400,
				Message: "Invalid Password",
			},
			wantError: true,
		},
		{
			name: "Invalid Username",
			credentials: schema.Credentials{
				Email:    "wrongUser@example.com",
				Password: "password1",
			},
			prepMongoMock: func(mt *mtest.T) {
				mt.AddMockResponses(bson.D{{Key: "ok", Value: 0}})
			},
			expectedError: schema.ErrorResponse{
				Code:    400,
				Message: "User does not exist!",
			},
			wantError: true,
		},
		{
			name: "Fail to Login Because Google User Already Exists",
			credentials: schema.Credentials{
				Email:    "user1@example.com",
				Password: "password1",
			},
			prepMongoMock: func(mt *mtest.T) {
				user := schema.User{
					Email:          "user1@example.com",
					Password:       "",
					SocialProvider: "GOOGLE",
				}

				userBSON, _ := bson.Marshal(user)
				var bsonD bson.D
				err := bson.Unmarshal(userBSON, &bsonD)
				if err != nil {
					assert.Fail("Failed to unmarshal bson data into document while prepping mock mongoDB. Method: 'Success'")
				}

				res := mtest.CreateCursorResponse(
					1,
					"data.users",
					mtest.FirstBatch,
					bsonD)
				end := mtest.CreateCursorResponse(
					0,
					"data.users",
					mtest.NextBatch)
				mt.AddMockResponses(res, end)
			},
			expectedError: schema.ErrorResponse{
				Code:    400,
				Message: "This email is already linked to GOOGLE",
			},
			wantError: true,
		},
	}

	// Create mock DB:
	opts := mtest.NewOptions().DatabaseName("data").ClientType(mtest.Mock)
	mt := mtest.New(t, opts)
	defer mt.Close()

	// For each test case:
	for _, testCase := range testCases {
		mt.Run(testCase.name, func(mt *mtest.T) {

			// Prep the mongo mocK:
			testCase.prepMongoMock(mt)

			loginResponse, err := Login(context, store, mt.Client, testCase.credentials)

			if testCase.wantError {
				if err == nil {
					assert.Fail("This test was supposed to throw an error!")
				} else {
					assert.Equal(testCase.expectedError.Code, err.Code)
					assert.Equal(testCase.expectedError.Message, err.Message)
				}
			} else {
				// Email is the same as expected
				assert.Equal(testCase.expectedResponse.User.Email, loginResponse.User.Email)
				// Password is returned as empty
				assert.Equal(loginResponse.User.Password, "")
				// CSRFToken returned:
				assert.True(loginResponse.CSRFToken != "")
			}
		})
	}
}

func TestCreateUserSession(t *testing.T) {
	assert := assert.New(t)

	// mocking the echo context
	e := echo.New()

	req := httptest.NewRequest(http.MethodPost, "/", nil)
	res := httptest.NewRecorder()
	c := e.NewContext(req, res)

	// initializing empty cookie store
	var store = sessions.NewCookieStore([]byte("test"))

	csrfToken, _ := CreateUserSession(c, store, "123")

	// 1. There is a csrfToken returned:
	assert.NotEqual(nil, csrfToken)

	// 2. There is a session created inside of the store that includes a CSRFToken and userId: 123
	session, _ := store.Get(c.Request(), "voxeti-session")

	assert.Equal("123", session.Values["userId"])
	assert.NotEqual(nil, session.Values["csrfToken"])
}

func TestInvalidateUserSession(t *testing.T) {
	assert := assert.New(t)

	// mocking the echo context
	e := echo.New()

	req := httptest.NewRequest(http.MethodPost, "/", nil)
	res := httptest.NewRecorder()
	c := e.NewContext(req, res)

	// initializing cookie store
	var store = sessions.NewCookieStore([]byte("test"))
	CreateUserSession(c, store, "123")

	// invalidate the user session
	InvalidateUserSession(c, store)

	// 1. Ensure the cookie is new when created and that no userId and CSRF token exists
	session, _ := store.Get(c.Request(), "voxeti-session")
	assert.Equal(nil, session.Values["userId"])
	assert.Equal(nil, session.Values["csrfToken"])
}

func TestAuthenticateSession(t *testing.T) {
	assert := assert.New(t)

	// mocking the echo context
	e := echo.New()
	e_unauthorized := echo.New()

	// initializing cookie store
	var store = sessions.NewCookieStore([]byte("test"))

	// Unauthorized user
	req_unauthorized := httptest.NewRequest(http.MethodPost, "/", nil)
	res_unauthorized := httptest.NewRecorder()
	context_unauthorized := e_unauthorized.NewContext(req_unauthorized, res_unauthorized)

	// Authorized user:
	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(nil))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	req.Header.Set("Csrftoken", "123")
	res := httptest.NewRecorder()
	c := e.NewContext(req, res)

	// create a cookie:
	session, _ := store.Get(req, "voxeti-session")
	session.Values["csrfToken"] = "123"
	session.Values["userId"] = "123"
	_ = session.Save(req, res)
	session.IsNew = false

	// 1. Authenticate the session with valid cookie and token:
	_, err := AuthenticateSession(c, store)
	if err != nil {
		assert.Fail(err.Message)
	}

	// 2. Check that AuthenticateSession fails when not provided a CSRFToken:
	req = httptest.NewRequest(http.MethodPost, "/", nil)
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	res = httptest.NewRecorder()
	c = e.NewContext(req, res)

	_, err = AuthenticateSession(c, store)

	assert.Equal(401, err.Code)

	// 3. Check that AuthenticateSession fails when csrfToken provided does not match cookie csrfToken:
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)

	session, _ = store.Get(req, "voxeti-session")
	session.Values["csrfToken"] = "something_else"
	session.Values["userId"] = "123"
	_ = session.Save(req, res)
	session.IsNew = false

	_, err = AuthenticateSession(c, store)

	assert.Equal(401, err.Code)

	// 4. Check that AuthenticateSession fails when cookie is expired:
	req = httptest.NewRequest(http.MethodPost, "/", nil)
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	res = httptest.NewRecorder()
	c = e.NewContext(req, res)

	session, _ = store.Get(req, "voxeti-session")
	session.Values["csrfToken"] = "123"
	session.Values["userId"] = "123"
	_ = session.Save(req, res)
	session.Options.MaxAge = -1
	session.IsNew = false

	_, err = AuthenticateSession(c, store)

	assert.Equal(401, err.Code)

	// 5. Should throw error when unauthorized user attempts to access current user's session state
	_, err = AuthenticateSession(context_unauthorized, store)

	assert.Equal(401, err.Code)
}

func TestCheckPasswordHash(t *testing.T) {
	assert := assert.New(t)

	input := "password1"
	invalidInput := "password2"
	hashedInput := "$2a$10$yQMzszWR14B7a8WmQh4GT.gf4bf/x1ntXpX0kobFKIW8kOHQ2DOji"

	// 1. Check that the method returns true for equivalent hashed, and unhashed strings:
	assert.True(CheckPasswordHash(input, hashedInput))

	// 2. Check that the method returns false of non-equivalent hashed, and unhashed strings:
	assert.True(!CheckPasswordHash(invalidInput, hashedInput))

	// 3. Catch edge case of small change made to Hashed password
	assert.True(!CheckPasswordHash("password1", "$2a$10$yQMzszWR14B7a8WmQh4GT.gf4bf/x1ntXpX0kobFKIW8kOHQ2DOj"))
}

func TestGenerateCSRFToken(t *testing.T) {
	assert := assert.New(t)

	csrfToken, err := GenerateCSRFToken()

	assert.True(err == nil)
	assert.True(csrfToken != "")
}

func TestValidateGoogleUser(t *testing.T) {
	assert := assert.New(t)

	// mocking the echo context
	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/", nil)
	res := httptest.NewRecorder()
	context := e.NewContext(req, res)
	// initializing cookie store
	var store = sessions.NewCookieStore([]byte("test"))
	testId, _ := primitive.ObjectIDFromHex("1")

	// Create DB Test Cases:
	testCases := []struct {
		name             string
		googleResponse   schema.GoogleResponse
		prepMongoMock    func(mt *mtest.T)
		expectedResponse schema.LoginResponse
		expectedError    schema.ErrorResponse
		wantNewUser      bool
		wantError        bool
	}{
		{
			name: "Success",
			googleResponse: schema.GoogleResponse{
				Email: "test@gmail.com",
				Scope: "",
			},
			prepMongoMock: func(mt *mtest.T) {
				user := schema.User{
					Id:             testId,
					Email:          "test@gmail.com",
					Password:       "",
					FirstName:      "Alder",
					LastName:       "Whiteford",
					SocialProvider: "GOOGLE",
				}

				userBSON, _ := bson.Marshal(user)
				var bsonD bson.D
				err := bson.Unmarshal(userBSON, &bsonD)
				if err != nil {
					assert.Fail("Failed to unmarshal bson data into document while prepping mock mongoDB. Method: 'Success'")
				}

				res := mtest.CreateCursorResponse(
					1,
					"data.users",
					mtest.FirstBatch,
					bsonD)
				end := mtest.CreateCursorResponse(
					0,
					"data.users",
					mtest.NextBatch)
				mt.AddMockResponses(res, end)
			},
			expectedResponse: schema.LoginResponse{
				User: schema.User{
					Id:             testId,
					Email:          "test@gmail.com",
					Password:       "",
					FirstName:      "Alder",
					LastName:       "Whiteford",
					SocialProvider: "GOOGLE",
				},
			},
			wantNewUser: false,
			wantError:   false,
		},
		{
			name: "No existing Google User",
			googleResponse: schema.GoogleResponse{
				Email: "test@gmail.com",
				Scope: "",
			},
			prepMongoMock: func(mt *mtest.T) {
				mt.AddMockResponses(bson.D{{Key: "ok", Value: 0}})
			},
			expectedResponse: schema.LoginResponse{
				User: schema.User{
					Email:          "test@gmail.com",
					Password:       "",
					FirstName:      "",
					LastName:       "",
					SocialProvider: "GOOGLE",
				},
			},
			wantNewUser: true,
			wantError:   false,
		},
		{
			name: "User is not a Google User",
			googleResponse: schema.GoogleResponse{
				Email: "test@gmail.com",
				Scope: "",
			},
			prepMongoMock: func(mt *mtest.T) {
				user := schema.User{
					Id:             testId,
					Email:          "test@gmail.com",
					Password:       "",
					FirstName:      "Alder",
					LastName:       "Whiteford",
					SocialProvider: "NONE",
				}

				userBSON, _ := bson.Marshal(user)
				var bsonD bson.D
				err := bson.Unmarshal(userBSON, &bsonD)
				if err != nil {
					assert.Fail("Failed to unmarshal bson data into document while prepping mock mongoDB. Method: 'Success'")
				}

				res := mtest.CreateCursorResponse(
					1,
					"data.users",
					mtest.FirstBatch,
					bsonD)
				end := mtest.CreateCursorResponse(
					0,
					"data.users",
					mtest.NextBatch)
				mt.AddMockResponses(res, end)
			},
			expectedError: schema.ErrorResponse{
				Code:    400,
				Message: "An account already exists with this email! Please re-attempt login with an email / password",
			},
			wantNewUser: false,
			wantError:   true,
		},
	}

	// Create mock DB:
	opts := mtest.NewOptions().DatabaseName("data").ClientType(mtest.Mock)
	mt := mtest.New(t, opts)
	defer mt.Close()

	// For each test case:
	for _, testCase := range testCases {
		mt.Run(testCase.name, func(mt *mtest.T) {

			// Prep the mongo mocK:
			testCase.prepMongoMock(mt)

			loginResponse, err := ValidateGoogleUser(context, store, &testCase.googleResponse, mt.Client)

			if testCase.wantError {
				if err == nil {
					assert.Fail("This test was supposed to throw an error!")
				} else {
					assert.Equal(testCase.expectedError.Code, err.Code)
					assert.Equal(testCase.expectedError.Message, err.Message)
				}
			} else {
				// Email is the same as expected
				assert.Equal(testCase.expectedResponse.User.Email, loginResponse.User.Email)
				// Password is returned as empty
				assert.Equal(loginResponse.User.Password, "")
				// Names are the same as expected
				assert.Equal(testCase.expectedResponse.User.FirstName, loginResponse.User.FirstName)
				assert.Equal(testCase.expectedResponse.User.LastName, loginResponse.User.LastName)

				if testCase.wantNewUser {
					// No CSRFToken if this is a new user:
					assert.Equal(loginResponse.CSRFToken, "")
				} else {
					// A CSRFToken should exist if it is an existing user:
					assert.NotEqual(loginResponse.CSRFToken, "")

					// Confirm a cookie has been created:
					session, _ := store.Get(context.Request(), "voxeti-session")

					assert.NotEqual(nil, session.Values["userId"])
					assert.NotEqual(nil, session.Values["csrfToken"])
				}
			}
		})
	}
}

func TestGetGoogleSSOUser(t *testing.T) {
	assert := assert.New(t)

	// Ensure the transaction will fail if provided an invalid access token:
	_, err := GetGoogleSSOUser(schema.GoogleAccessToken{AccessToken: "123"})

	assert.Equal(err.Code, 400)
	assert.Equal(err.Message, "Invalid access token!")
}
