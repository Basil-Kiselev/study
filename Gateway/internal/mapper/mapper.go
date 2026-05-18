package mapper

import (
	"study/Gateway/internal/model"
	accountpb "study/contracts/account"
	authpb "study/contracts/auth"
	"study/contracts/gateway"

	"google.golang.org/protobuf/types/known/timestamppb"
)

func PbToUser(userpb *accountpb.User) model.User {
	return model.User{
		ID:         userpb.Id,
		Login:      userpb.Login,
		Email:      userpb.Email,
		Phone:      userpb.Phone,
		FirstName:  userpb.FirstName,
		LastName:   userpb.LastName,
		MiddleName: userpb.MiddleName,
		Age:        userpb.Age,
		CreatedAt:  userpb.CreatedAt.AsTime(),
		UpdatedAt:  userpb.UpdatedAt.AsTime(),
	}
}

func UserToPb(user model.User) *accountpb.User {
	return &accountpb.User{
		Id:         user.ID,
		Login:      user.Login,
		Email:      user.Email,
		Phone:      user.Phone,
		FirstName:  user.FirstName,
		LastName:   user.LastName,
		MiddleName: user.MiddleName,
		Age:        user.Age,
		CreatedAt:  timestamppb.New(user.CreatedAt),
		UpdatedAt:  timestamppb.New(user.UpdatedAt),
	}
}

func UsersToPb(users []model.User) []*accountpb.User {
	result := make([]*accountpb.User, len(users))
	for i, user := range users {
		result[i] = UserToPb(user)
	}
	return result
}

func PbToUserCreate(accountpbUser *accountpb.CreateUser) model.CreateUser {
	return model.CreateUser{
		Login:      accountpbUser.Login,
		Email:      accountpbUser.Email,
		Phone:      accountpbUser.Phone,
		FirstName:  accountpbUser.FirstName,
		LastName:   accountpbUser.LastName,
		MiddleName: accountpbUser.MiddleName,
		Age:        accountpbUser.Age,
	}
}

func PbToUserUpdate(accountpbUser *accountpb.User) model.UpdateUser {
	return model.UpdateUser{
		Email:      accountpbUser.Email,
		Phone:      accountpbUser.Phone,
		FirstName:  accountpbUser.FirstName,
		LastName:   accountpbUser.LastName,
		MiddleName: accountpbUser.MiddleName,
		Age:        accountpbUser.Age,
	}
}

func PbToRegisterUser(regpb *authpb.RegisterRequest) model.RegisterUser {
	return model.RegisterUser{
		Login:    regpb.Login,
		Email:    regpb.Email,
		Password: regpb.Password,
	}
}

func PbToLoginUser(loginpb *gateway.LoginRequest) model.LoginUser {
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
