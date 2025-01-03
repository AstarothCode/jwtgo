package v1

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"

	"jwtgo/internal/app/controller/http/dto"
	"jwtgo/internal/app/controller/http/mapper"
	"jwtgo/internal/app/controller/http/middleware"
	customErr "jwtgo/internal/app/error"
	serviceInterface "jwtgo/internal/app/interface/service"
	"jwtgo/internal/pkg/request"
	"jwtgo/internal/pkg/request/schema"
	"jwtgo/pkg/logging"
)

type AuthController struct {
	authService      serviceInterface.AuthService
	requestValidator *validator.Validate
	logger           *logging.Logger
}

func NewAuthController(
	authService serviceInterface.AuthService,
	requestValidator *validator.Validate,
	logger *logging.Logger,
) *AuthController {
	return &AuthController{
		authService:      authService,
		requestValidator: requestValidator,
		logger:           logger,
	}
}

func (ac *AuthController) Register(router *gin.Engine) {
	router.POST("/auth/signup", middleware.Validator[dto.UserCredentialsDTO](ac.requestValidator), ac.SignUp())
	router.POST("/auth/signin", middleware.Validator[dto.UserCredentialsDTO](ac.requestValidator), ac.SignIn())
	router.POST("/auth/refresh", ac.Refresh())
}

func (ac *AuthController) SignUp() gin.HandlerFunc {
	return func(c *gin.Context) {
		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()

		userCredentialsDTO := c.MustGet("validatedBody").(dto.UserCredentialsDTO)

		_, err := ac.authService.SignUp(ctx, &userCredentialsDTO)
		if err != nil {
			var alreadyExistsErr *customErr.AlreadyExistsError

			if errors.As(err, &alreadyExistsErr) {
				c.JSON(http.StatusConflict, gin.H{"message": err.Error()})
			} else {
				ac.logger.Error("Error while authorizing: ", err)
				c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
			}

			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "User successfully registered"})
	}
}

func (ac *AuthController) SignIn() gin.HandlerFunc {
	return func(c *gin.Context) {
		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()

		userCredentialsDTO := c.MustGet("validatedBody").(dto.UserCredentialsDTO)

		userTokensDTO, err := ac.authService.SignIn(ctx, &userCredentialsDTO)
		if err != nil {
			var invalidCredentialsErr *customErr.InvalidCredentialsError

			if errors.As(err, &invalidCredentialsErr) {
				c.JSON(http.StatusUnauthorized, gin.H{"message": err.Error()})
			} else {
				ac.logger.Error("Error while registering: ", err)
				c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
			}

			return
		}

		request.SetCookies(c, []schema.Cookie{
			{Name: "access_token", Value: userTokensDTO.AccessToken, Duration: 7 * 24 * time.Hour},
			{Name: "refresh_token", Value: userTokensDTO.RefreshToken, Duration: 7 * 24 * time.Hour},
		})

		c.JSON(http.StatusOK, gin.H{"message": "Logged in successfully"})
	}
}

func (ac *AuthController) Refresh() gin.HandlerFunc {
	return func(c *gin.Context) {
		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()

		refreshToken, err := c.Cookie("refresh_token")
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid refresh token"})
			return
		}

		refreshTokenDTO := mapper.MapToUserRefreshTokenDTO(refreshToken)

		userTokensDTO, err := ac.authService.Refresh(ctx, refreshTokenDTO)
		if err != nil {
			var invalidTokenError *customErr.InvalidTokenError
			var expiredTokenError *customErr.ExpiredTokenError
			var userNotFoundError *customErr.UserNotFoundError

			if errors.As(err, &invalidTokenError) || errors.As(err, &expiredTokenError) || errors.As(err, &userNotFoundError) {
				c.JSON(http.StatusUnauthorized, gin.H{"message": err.Error()})
			} else {
				ac.logger.Error("Error while refreshing: ", err)
				c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
			}

			return
		}

		request.SetCookies(c, []schema.Cookie{
			{Name: "access_token", Value: userTokensDTO.AccessToken, Duration: 7 * 24 * time.Hour},
			{Name: "refresh_token", Value: userTokensDTO.RefreshToken, Duration: 7 * 24 * time.Hour},
		})

		c.JSON(http.StatusOK, gin.H{"message": "Tokens updated successfully"})
	}
}
