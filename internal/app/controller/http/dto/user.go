package dto

type UserRefreshTokenDTO struct {
	RefreshToken string `json:"refresh_token"`
}

type UserTokensDTO struct {
	AccessToken  string `json:"accessToken"`
	RefreshToken string `json:"refresh_token"`
}

type UserCredentialsDTO struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=6,max=64"`
}
