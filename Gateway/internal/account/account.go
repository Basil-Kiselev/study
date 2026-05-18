package account

import (
	"context"
	"fmt"
	"study/Gateway/internal/model"
	accountpb "study/contracts/account"
	"study/contracts/pagination"
)

type Service struct {
	client accountpb.AccountClient
}

func NewService(client accountpb.AccountClient) *Service {
	return &Service{
		client: client,
	}
}

func (s *Service) GetUser(ctx context.Context, userID uint64) (model.User, error) {
	user, err := s.client.GetUser(ctx, &accountpb.GetUserRequest{UserId: userID})
	if err != nil {
		return model.User{}, fmt.Errorf("fail to get user: %w", err)
	}
	return PbToUser(user.User), nil
}

func (s *Service) GetUsers(ctx context.Context, limit uint32, offset uint32) ([]model.User, error) {
	users, err := s.client.GetUsers(ctx, &accountpb.GetUsersRequest{Pagination: &pagination.Pagination{Offset: offset, Limit: limit}})
	if err != nil {
		return []model.User{}, fmt.Errorf("fail to get users: %w", err)
	}
	return PbToUsers(users.Users), nil
}

func (s *Service) DeleteUser(ctx context.Context, userID uint64) error {
	_, err := s.client.DeleteUser(ctx, &accountpb.DeleteUserRequest{UserId: userID})
	if err != nil {
		return fmt.Errorf("fail to delete user: %w", err)
	}

	return nil
}

func (s *Service) CreateUser(ctx context.Context, user model.CreateUser) (model.User, error) {
	resp, err := s.client.CreateUser(ctx, &accountpb.CreateUserRequest{User: UserCreateToPb(user)})
	if err != nil {
		return model.User{}, fmt.Errorf("fail to create user: %w", err)
	}

	return PbToUser(resp.User), nil
}

func (s *Service) UpdateUser(ctx context.Context, userID uint64, user model.UpdateUser) error {
	_, err := s.client.UpdateUser(ctx, &accountpb.UpdateUserRequest{UserId: userID, User: UserToPb(user)})
	if err != nil {
		return fmt.Errorf("fail to update user: %w", err)
	}

	return nil
}

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

func UserToPb(user model.UpdateUser) *accountpb.User {
	return &accountpb.User{
		Email:      user.Email,
		Phone:      user.Phone,
		FirstName:  user.FirstName,
		LastName:   user.LastName,
		MiddleName: user.MiddleName,
		Age:        user.Age,
	}
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

func PbToUsers(pbUsers []*accountpb.User) []model.User {
	res := make([]model.User, 0, len(pbUsers))
	for _, user := range pbUsers {
		user := PbToUser(user)
		res = append(res, user)
	}
	return res
}

func UserCreateToPb(user model.CreateUser) *accountpb.CreateUser {
	return &accountpb.CreateUser{
		Login:      user.Login,
		Email:      user.Email,
		LastName:   user.LastName,
		FirstName:  user.FirstName,
		Phone:      user.Phone,
		MiddleName: user.MiddleName,
		Age:        user.Age,
	}
}
