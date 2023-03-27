package util

import (
	board "auth-service/boardgRPCoperate"
	"auth-service/models"
	"context"
	"log"
)

type BoardServer struct {
	board.UnimplementedBoardServiceServer
	MongoModel models.Mongo
}

func (b *BoardServer) LikeComment(ctx context.Context, info *board.LikeCommentInfo) (*board.Response, error) {
	log.Println(info.GetGiverID(), info.GetReceiverID(), info.GetCommentID(), info.GetIsGiverLike())
	err := b.MongoModel.LikeComment(info.GetGiverID(), info.GetReceiverID(), info.GetCommentID(), info.GetIsGiverLike())
	if err != nil {
		log.Println(err)
		res := &board.Response{
			Result: false,
		}
		return res, err
	}
	res := &board.Response{
		Result: true,
	}
	return res, nil
}
func (b *BoardServer) ReleaseComment(ctx context.Context, info *board.ReleaseCommentInfo) (*board.Response, error) {
	log.Println(info.GetUserID(), info.GetCommentID())
	err := b.MongoModel.ReleaseComment(info.GetUserID(), info.GetCommentID())
	if err != nil {
		log.Println(err)
		res := &board.Response{
			Result: false,
		}
		return res, err
	}
	res := &board.Response{
		Result: true,
	}
	return res, nil
}
func (b *BoardServer) CollectBoard(ctx context.Context, info *board.CollectBoardInfo) (*board.Response, error) {
	err := b.MongoModel.CollectBoard(info.GetUserID(), info.GetBoardID(), info.GetBoardTitle(), info.GetIsCollect())
	if err != nil {
		log.Println(err)
		res := &board.Response{
			Result: false,
		}
		return res, err
	}
	res := &board.Response{
		Result: true,
	}
	return res, nil
}
func (b *BoardServer) LikeBoard(ctx context.Context, info *board.LikeBoardInfo) (*board.Response, error) {
	err := b.MongoModel.LikedBoard(info.GetUserID(), info.GetBoardID(), info.GetIsLiked())
	if err != nil {
		log.Println(err)
		res := &board.Response{
			Result: false,
		}
		return res, err
	}
	res := &board.Response{
		Result: true,
	}
	return res, nil
}
