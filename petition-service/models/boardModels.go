package models

import (
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
type Boards struct {
	ID              bson.ObjectId `bson:"_id"`
	Date            time.Time     `bson:"date"`
	BoardTitle      string        `bson:"board-title"`
	BoardMotivation string        `bson:"board-motivation"`
	NumberOfSigners int           `bson:"number-of-signers"`
	Signers         []string      `bson:"signers"`
}
type SigningBoards struct {
	ID              bson.ObjectId
	BoardTitle      string
	NumberOfSigners int
	IsUserSupport   bool
}
type AchievedBoards struct {
	ID         bson.ObjectId
	BoardTitle string
}
type AllBoards struct {
	SigningBoard  []*SigningBoards  `json:"signing-board"`
	AchievedBoard []*AchievedBoards `json:"recently-achieved-board"`
}

func NewMongo() (*Mongo, error) {
	session, err := mgo.Dial("mongodb://mongo_petition:27017")
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
		collection: session.DB("board").C("BoardStatus"),
	}
	// Do something with the database
	fmt.Println("Connected to MongoDB!")

	return m, nil
}
func (m *Mongo) GetAllBoard(userID string) (*AllBoards, error) {
	var boards []Boards
	var all AllBoards
	fromDate := time.Date(time.Now().Year(), time.Now().Month(), time.Now().Day()-6, 0, 0, 0, 0, time.UTC)
	tomorrow := time.Date(time.Now().Year(), time.Now().Month(), time.Now().Day()+1, 0, 0, 0, 0, time.UTC)
	log.Println(fromDate,tomorrow)
	err := m.collection.Find(
		bson.M{
			"date": bson.M{
				"$gte": fromDate,
				"$lt": tomorrow,
		},
	}).All(&boards)
	//err:=m.collection.Find(nil).All(&boards)
	log.Println(boards)
	if err != nil {
		return nil, err
	}
	// if userID == "" {
	// 	log.Println(boards)
	// 	return &boards, nil
	// }
	for _, t := range boards {
		if t.NumberOfSigners > 30 {
			all.AchievedBoard = append(all.AchievedBoard, &AchievedBoards{
				ID:         t.ID,
				BoardTitle: t.BoardTitle,
			})
		} else {
			flag := false
			for _, j := range t.Signers {
				if j == userID {
					flag = true
					break
				}
			}
			all.SigningBoard = append(all.SigningBoard, &SigningBoards{
				ID:              t.ID,
				BoardTitle:      t.BoardTitle,
				NumberOfSigners: t.NumberOfSigners,
				IsUserSupport:   flag,
			})
		}
	}
	return &all, nil
}
func (m *Mongo) SupportBoard(UserID, BoardID string, isSupport bool) error {
	var update bson.M
	if isSupport {
		update = bson.M{
			"$inc":      bson.M{"number-of-signers": 1},
			"$addToSet": &bson.M{"signers": UserID}}
	} else {
		update = bson.M{
			"$inc":  bson.M{"number-of-signers": -1},
			"$pull": &bson.M{"signers": UserID}}
	}
	err := m.collection.Update(bson.M{"_id": bson.ObjectIdHex(BoardID)}, update)
	if err != nil {
		return err
	}
	return nil
}
func (m *Mongo) InitBoard(BoardTitle, BoardMotivation string) error {
	err := m.collection.Insert(&Boards{
		ID: bson.NewObjectId(),
		Date:            time.Now().UTC(),
		BoardTitle:      BoardTitle,
		BoardMotivation: BoardMotivation,
		NumberOfSigners: 0,
		Signers:         []string{},
	})
	if err != nil {
		log.Println(err)
		return err
	}
	return nil
}
func (m *Mongo) GetBoard(boardID string) (*Boards, error) {
	var result Boards
	if err := m.collection.Find(bson.M{"_id": bson.ObjectIdHex(boardID)}).One(&result); err != nil {
		log.Println(err)
		return nil, err
	}
	return &result, nil
}
