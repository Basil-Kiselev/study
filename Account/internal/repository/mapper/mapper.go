package mapper

import (
	"study/Account/internal/model"
	repmodel "study/Account/internal/repository/model"
)

func UserToRepoUser(user model.User) repmodel.User {
	return repmodel.User{
		ID:         user.ID,
		Login:      user.Login,
		Email:      user.Email,
		Phone:      user.Phone,
		FirstName:  user.FirstName,
		LastName:   user.LastName,
		MiddleName: user.MiddleName,
		Age:        user.Age,
		CreatedAt:  user.CreatedAt,
		UpdatedAt:  user.UpdatedAt,
		Balance:    user.Balance,
		IsDeleted:  user.IsDeleted,
	}
}

func RepoUserToUser(user repmodel.User) model.User {
	return model.User{
		ID:         user.ID,
		Login:      user.Login,
		Email:      user.Email,
		Phone:      user.Phone,
		FirstName:  user.FirstName,
		LastName:   user.LastName,
		MiddleName: user.MiddleName,
		Age:        user.Age,
		CreatedAt:  user.CreatedAt,
		UpdatedAt:  user.UpdatedAt,
		Balance:    user.Balance,
		IsDeleted:  user.IsDeleted,
	}
}

func RepoUsersToUsers(users []repmodel.User) []model.User {
	res := make([]model.User, len(users))
	for i, user := range users {
		res[i] = RepoUserToUser(user)
	}
	return res
}

func UpdateUserToRepoUser(user model.User) repmodel.User {
	return repmodel.User{
		Email:      user.Email,
		Phone:      user.Phone,
		FirstName:  user.FirstName,
		LastName:   user.LastName,
		MiddleName: user.MiddleName,
		Age:        user.Age,
		Balance:    user.Balance,
		IsDeleted:  user.IsDeleted,
	}
}
