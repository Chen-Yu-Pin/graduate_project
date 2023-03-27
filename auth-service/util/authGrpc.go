package util

import (
	"auth-service/auth"
	"auth-service/models"
	"context"
	"log"
)

type AuthServer struct {
	auth.UnimplementedAuthServiceServer
	RedisModel models.Redis
}

func (a *AuthServer) AuthAccount(ctx context.Context, req *auth.AuthRequest) (*auth.AuthResponse, error) {
	input := req.GetAuthEntry()
	err := a.RedisModel.SearchToken(input.UserID, input.Token)
	if err != nil {
		log.Println(err)
		res := &auth.AuthResponse{
			Result: false,
		}
		return res, err
	}
	err = ParseToken(input.Token)
	if err != nil {
		log.Println(err)
		res := &auth.AuthResponse{
			Result: false,
		}
		return res, err
	}
	res := &auth.AuthResponse{
		Result: true,
	}
	return res, nil
}
