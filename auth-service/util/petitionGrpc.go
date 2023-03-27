package util

import (
	"auth-service/models"
	petition "auth-service/petitiongRPC"
	"context"
)

type PetitionServer struct {
	petition.UnimplementedPetitionServiceServer
	MongoModel models.Mongo
}

func (p *PetitionServer) AddSigningBoard(ctx context.Context, info *petition.SignBoard) (*petition.Response, error) {
	err := p.MongoModel.SignBoard(info.GetUserID(), info.GetSignBoardID(), info.GetIsSign())
	if err != nil {
		return &petition.Response{
			Result: false,
		}, err
	}
	return &petition.Response{
		Result: true,
	}, nil
}

func (p *PetitionServer) LaunchPetitionBoard(ctx context.Context, info *petition.LaunchBoard) (*petition.Response, error) {
	err := p.MongoModel.LaunchBoard(info.GetUserID())
	if err != nil {
		return &petition.Response{
			Result: false,
		}, err
	}
	return &petition.Response{
		Result: true,
	}, nil
}
