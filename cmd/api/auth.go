package main

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/nagakushal786/post-ur-world/internal/mailer"
	"github.com/nagakushal786/post-ur-world/internal/store"
)

type RegisterUserPayload struct{
	Username string `json:"username" validate:"required,max=100"`
	Email string `json:"email" validate:"required,email,max=255"`
	Password string `json:"password" validate:"required,min=3,max=72"`
}

type UserWithToken struct{
	*store.User
	Token string `json:"token"`
}

// RegisterUser godoc
//
// @Summary Registers a user
// @Description Registers a user
// @Tags authentication
// @Accept json
// @Produce json
// @Param payload body RegisterUserPayload true "User Credentials"
// @Success 201 {object} UserWithToken "User registered"
// @Failure 400 {object} error
// @Failure 500 {object} error
// @Router /authentication/register [post]
func (app *application) registerUserHandler(w http.ResponseWriter, req *http.Request){
	var payload RegisterUserPayload
	if err:=readJSON(w, req, &payload); err!=nil{
		app.badRequestError(w, req, err)
		return
	}

	if err:=Validate.Struct(payload); err!=nil{
		app.badRequestError(w, req, err)
		return
	}

	user:=&store.User{
		Username: payload.Username,
		Email: payload.Email,
		Role: store.Role{
			Name: "user",
		},
	}

	if err:=user.Password.Set(payload.Password); err!=nil{
		app.internalServerError(w, req, err)
		return
	}

	ctx:=req.Context()

	plainToken:=uuid.New().String()

	// For storing
	hash:=sha256.Sum256([]byte(plainToken))
	hashToken:=hex.EncodeToString(hash[:])

	err:=app.store.Users.CreateAndInvite(ctx, user, hashToken, app.config.mail.exp)
	if err!=nil{
		switch err{
			case store.ErrDuplicateEmail:
				app.badRequestError(w, req, err)
			case store.ErrDuplicateUsername:
				app.badRequestError(w, req, err)
			default:
				app.internalServerError(w, req, err)
		}
		return
	}

	userWithToken:=UserWithToken{
		User: user,
		Token: plainToken,
	}

	activationURL:=fmt.Sprintf("%s/confirm/%s", app.config.frontendURL, plainToken)

	vars:=struct{
		Username string
		ActivationURL string
	}{
		Username: user.Username,
		ActivationURL: activationURL,
	}

	// send email

	// For development
	isSandbox:=false
	status, err:=app.mailer.Send(mailer.UserWelcomeTemplate, user.Username, user.Email, vars, isSandbox)

	// For production by setting env to "production"
	// IsProdEnv:=app.config.env=="production"
	// status, err:=app.mailer.Send(mailer.UserWelcomeTemplate, user.Username, user.Email, vars, !IsProdEnv)
	if err!=nil{
		app.logger.Errorf("Error sending invitation mail", "error", err)

		// rollback user creation if email fails (SAGA pattern)
		if err:=app.store.Users.Delete(ctx, user.ID); err!=nil{
			app.logger.Errorw("Error deleting user", "error", err)
		}
	}

	app.logger.Infow("Email sent successfully", "status code", status)

	if err:=app.jsonResponse(w, http.StatusCreated, userWithToken); err!=nil{
		app.internalServerError(w, req, err)
	}
}

type CreateUserTokenPayload struct{
	Email string `json:"email" validate:"required,email,max=255"`
	Password string `json:"password" validate:"required,min=3,max=72"`
}

// CreateToken godoc
//
// @Summary Creates a token
// @Description Creates a token
// @Tags authentication
// @Accept json
// @Produce json
// @Param payload body CreateUserTokenPayload true "User Credentials"
// @Success 200 {object} string "Token created"
// @Failure 400 {object} error
// @Failure 401 {object} error
// @Failure 500 {object} error
// @Router /authentication/token [post]
func (app *application) createTokenHandler(w http.ResponseWriter, req *http.Request){
	// parse payload credentials
	var payload CreateUserTokenPayload
	if err:=readJSON(w, req, &payload); err!=nil{
		app.badRequestError(w, req, err)
		return
	}

	if err:=Validate.Struct(payload); err!=nil{
		app.badRequestError(w, req, err)
		return
	}

	// fetch the user from credentials
	user, err:=app.store.Users.GetByEmail(req.Context(), payload.Email)
	if err!=nil{
		switch err{
			case store.ErrNotFound:
				app.unAuthorizationError(w, req, err)
			default:
				app.internalServerError(w, req, err)
		}
		return
	}

	// generate token -> add claims
	claims:=jwt.MapClaims{
		"sub": user.ID,
		"exp": time.Now().Add(app.config.auth.token.exp).Unix(),
		"iat": time.Now().Unix(),
		"nbf": time.Now().Unix(),
		"iss": app.config.auth.token.issuer,
		"aud": app.config.auth.token.audience,
	}

	token, err:=app.authenticator.GenerateToken(claims)
	if err!=nil{
		app.internalServerError(w, req, err)
		return
	}

	// send it to the client
	if err:=app.jsonResponse(w, http.StatusCreated, token); err!=nil{
		app.internalServerError(w, req, err)
	}
}