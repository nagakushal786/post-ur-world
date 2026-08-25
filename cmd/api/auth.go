package main

import (
	"crypto/sha256"
	"encoding/hex"
	"net/http"

	"github.com/google/uuid"
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

	if err:=app.jsonResponse(w, http.StatusCreated, userWithToken); err!=nil{
		app.internalServerError(w, req, err)
	}
}