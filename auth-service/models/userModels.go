package models

import (
	"errors"
	"fmt"
	"log"
	"os"
	"time"

	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

type Mongo struct {
	session    *mgo.Session
	collection *mgo.Collection
}

type UserInfo struct {
	UserEstablishedDate    time.Time                `bson:"user-established-date" json:"user-established-date"`
	UserID                 bson.ObjectId            `bson:"_id" json:"user-id"`
	UserName               string                   `bson:"user-name" json:"user-name"`
	UserPassword           string                   `bson:"user-password" json:"user-password"`
	UserHeadshotNumber     string                   `bson:"user-headshot-number" json:"user-headshot-number"`
	UserMail               string                   `bson:"user-mail" json:"user-mail"`
	PersonalOperateHistory PersonalOperateHistorys  `bson:"personal-operate-history" json:"personal-operate-history"`
	CollectedBoard         []CollectedBoards        `bson:"collected-boards" json:"collected-boards"`
	LikededBoard           []LikedBoards            `bson:"liked-boards" json:"liked-boards"`
	LikedComment           []LikedComments          `bson:"liked-comments" json:"liked-Comments"`
	PostedComment          []PostedComments         `bson:"posted-comments" json:"posted-comments"`
	SupportedSigningBoard  []SupportedSigningBoards `bson:"supported-signing-boards" json:"supported-signing-boards"`
}
type PersonalOperateHistorys struct {
	NumberOfReceivedLikes        int `bson:"number-of-received-likes" json:"number-of-received-likes"`
	NumberOfReleasedComments     int `bson:"number-of-released-comments" json:"number-of-released-comments"`
	NumberOfCollectBoards        int `bson:"number-of-collect-boards" json:"number-of-collect-boards"`
	NumberOfLikeBoards           int `bson:"number-of-like-boards" json:"number-of-like-boards"`
	NumberOfLikeComments         int `bson:"number-of-like-comments" json:"number-of-like-comments"`
	NumberOfLaunchPetitionBoards int `bson:"number-of-launch-petition-boards" json:"number-of-launch-petition-boards"`
	NumberOfSupportSigningBoards int `bson:"number-of-support-signing-boards" json:"number-of-support-signing-boards"`
}
type CollectedBoards struct {
	BoardID    string `bson:"board-id" json:"board-id"`
	BoardTitle string `bson:"board-title" json:"board-title"`
}
type LikedBoards struct {
	BoardID string `bson:"board-id" json:"board-id"`
}
type LikedComments struct {
	CommentID string `bson:"comment-id" json:"comment-id"`
}
type PostedComments struct {
	CommentID string `bson:"comment-id" json:"comment-id"`
}
type SupportedSigningBoards struct {
	SigningBoardID string `bson:"signing-board-id" json:"signing-board-id"`
}

func NewMongo() (*Mongo, error) {
	session, err := mgo.Dial("mongodb://mongo_user:27017")
	if err != nil {
		fmt.Println("Error connecting to MongoDB:", err)
		return nil, err
	}
	cred := &mgo.Credential{
		Username: os.Getenv("user"),
		Password: os.Getenv("password"),
	}
	err = session.Login(cred)
	if err != nil {
		panic(err)
	}

	session.SetMode(mgo.Monotonic, true)
	m := &Mongo{
		session:    session,
		collection: session.DB("user").C("userinfo"),
	}
	for _, key := range []string{"user-name", "user-mail"} {
		index := mgo.Index{
			Key:    []string{key},
			Unique: true,
		}
		if err := m.collection.EnsureIndex(index); err != nil {
			panic(err)
		}
	}
	// Do something with the database
	fmt.Println("Connected to MongoDB!")

	return m, nil
}

func (m *Mongo) Close() {
	m.session.Close()
}

func (m *Mongo) InsertUser(user *UserInfo) error {

	if err := m.collection.Insert(user); err != nil {
		log.Println(err)
		return err
	}
	return nil
}

func (m *Mongo) LoginCheck(UserAccount, UserPassword string) (bool, *UserInfo, error) {

	var account UserInfo
	query := bson.M{
		"$and": []bson.M{
			{
				"$or": []bson.M{
					{"user-name": UserAccount},
					{"user-mail": UserAccount},
				},
			},
			{"user-password": UserPassword},
		},
	}
	err := m.collection.Find(&query).One(&account)
	log.Println("the result: ", account)
	if err != nil || account.UserID == "" {
		log.Println(err)
		return false, nil, err
	}
	if account.UserID == "" {

		return false, nil, errors.New("error account or password")
	}
	return true, &account, nil

}

func (m *Mongo) SearchUser(userID string) (*UserInfo, error) {
	var person UserInfo
	if err := m.collection.Find(bson.M{"_id": bson.ObjectIdHex(userID)}).One(&person); err != nil {
		log.Println(err)
		return nil, err
	}
	return &person, nil
}

func (m *Mongo) LikeComment(giver, receiver, commentID string, isLike bool) error {
	var update []bson.M
	if isLike {
		update = append(update,
			bson.M{
				"$inc": &bson.M{"personal-operate-history.number-of-like-comments": 1},
				"$addToSet": &bson.M{"liked-comments": &LikedComments{
					CommentID: commentID,
				}},
			})
		update = append(update,
			bson.M{
				"$inc": &bson.M{"personal-operate-history.number-of-received-likes": 1},
			})
	} else {
		update = append(update,
			bson.M{
				"$inc": &bson.M{"personal-operate-history.number-of-like-comments": -1},
				"$pull": &bson.M{"liked-comments": &LikedComments{
					CommentID: commentID,
				}},
			})
		update = append(update,
			bson.M{
				"$inc": &bson.M{"personal-operate-history.number-of-received-likes": -1},
			})
	}
	err := m.collection.Update(bson.M{"_id": bson.ObjectIdHex(giver)}, update[0])
	if err != nil {
		return err
	}
	err = m.collection.Update(bson.M{"_id": bson.ObjectIdHex(receiver)}, update[1])
	if err != nil {
		return err
	}
	return nil
}
func (m *Mongo) ReleaseComment(userID, commentID string) error {
	log.Println("in userModels releaseComment", userID, commentID)
	err := m.collection.Update(bson.M{"_id": bson.ObjectIdHex(userID)}, bson.M{"$inc": bson.M{"personal-operate-history.number-of-released-comments": 1}})
	if err != nil {
		return err
	}

	err = m.collection.Update(bson.M{"_id": bson.ObjectIdHex(userID)},
		bson.M{"$addToSet": bson.M{"posted-comments": &PostedComments{
			CommentID: commentID,
		}}})
	if err != nil {
		return err
	}
	log.Println("in userModels releaseComment finish")
	return nil
}

func (m *Mongo) CollectBoard(UserID, BoardID, BoardTitle string, isCollect bool) error {
	var update = bson.M{}
	if isCollect {
		update = bson.M{
			"$inc": bson.M{"personal-operate-history.number-of-collect-boards": 1},
			"$addToSet": &bson.M{"collected-boards": &CollectedBoards{
				BoardID:    BoardID,
				BoardTitle: BoardTitle,
			}}}
	} else {
		update = bson.M{
			"$inc": bson.M{"personal-operate-history.number-of-collect-boards": -1},
			"$pull": &bson.M{"collected-boards": &CollectedBoards{
				BoardID:    BoardID,
				BoardTitle: BoardTitle,
			}}}
	}
	err := m.collection.Update(bson.M{"_id": bson.ObjectIdHex(UserID)}, update)
	if err != nil {
		return err
	}
	return nil
}

func (m *Mongo) LikedBoard(userID, boardID string, isLike bool) error {
	var update = bson.M{}
	if isLike {
		update = bson.M{
			"$inc": bson.M{"personal-operate-history.number-of-like-boards": 1},
			"$addToSet": bson.M{"liked-boards": &LikedBoards{
				BoardID: boardID,
			}}}
	} else {
		update = bson.M{
			"$inc": bson.M{"personal-operate-history.number-of-like-boards": -1},
			"$pull": bson.M{"liked-boards": &LikedBoards{
				BoardID: boardID,
			}}}
	}
	err := m.collection.Update(bson.M{"_id": bson.ObjectIdHex(userID)}, update)
	if err != nil {
		return err
	}
	return nil
}

func (m *Mongo) SignBoard(userID, signBoardID string, isSign bool) error {
	var update = bson.M{}
	if isSign {
		update = bson.M{
			"$inc": bson.M{"personal-operate-history.number-of-support-signing-boards": 1},
			"$addToSet": bson.M{"supported-signing-boards": &SupportedSigningBoards{
				SigningBoardID: signBoardID,
			}}}
	} else {
		update = bson.M{
			"$inc": bson.M{"personal-operate-history.number-of-support-signing-boards": -1},
			"$pull": bson.M{"supported-signing-boards": &SupportedSigningBoards{
				SigningBoardID: signBoardID,
			}}}
	}
	err := m.collection.Update(bson.M{"_id": bson.ObjectIdHex(userID)}, update)
	if err != nil {
		return err
	}
	return nil
}
func (m *Mongo) LaunchBoard(userID string) error {
	err := m.collection.Update(bson.M{"_id": bson.ObjectIdHex(userID)}, bson.M{"$inc": bson.M{"personal-operate-history.number-of-launch-petition-boards": 1}})
	if err != nil {
		return err
	}
	return nil
}
