package mapper

import (
	"study/Auth/internal/model"
	authpb "study/contracts/auth"
)

func PbToRegisterUser(regpb *authpb.RegisterRequest) model.RegisterUser {
	return model.RegisterUser{
		Login:    regpb.Login,
		Email:    regpb.Email,
		Password: regpb.Password,
	}
}

func PbToLoginUser(loginpb *authpb.LoginRequest) model.LoginUser {
	return model.LoginUser{
		LoginOrMail: loginpb.LoginOrEmail,
		Password:    loginpb.Password,
	}
}

func TokenPairTopb(tokenPair model.TokenPair) *authpb.TokenPair {
	return &authpb.TokenPair{
		AccessToken:  tokenPair.AccessToken,
		RefreshToken: tokenPair.RefreshToken,
	}
}
