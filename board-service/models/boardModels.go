package models

import (
	"log"
	"time"

	"github.com/lib/pq"
	uuid "github.com/satori/go.uuid"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type DataModel struct {
	db *gorm.DB
}

type Board struct {
	ID                 uuid.UUID           `gorm:"primary_key;type:uuid;default:gen_random_uuid()" json:"id"`
	Date               time.Time           `json:"date"`
	Title              string              `json:"title"`
	IsUserCollected    bool                `gorm:"-:all" json:"is-user-collected"`
	IsUserLike         bool                `gorm:"-:all" json:"is-user-like"`
	ImageCloud         string              `json:"image-cloud"`
	SourceLinks        []SourceLink        `json:"source-link-list"`
	Tags               []Tag               `gorm:"many2many:board_tags;" json:"tags"`
	KeywordsWithCounts []KeywordsWithCount `json:"keywords-with-count"`
	Positions          []Position          `json:"positions"`
	NumOfLike          int                 `json:"number-of-like"`
	Liker              pq.StringArray      `gorm:"type:text[]" json:"-"`
	Collecter          pq.StringArray      `gorm:"type:text[]" json:"-"`
	Count              uint                `json:"-"`
}

type Tag struct {
	ID uuid.UUID `gorm:"primary_key;type:uuid;default:gen_random_uuid()" json:"-"`
	//BoardID uuid.UUID `gorm:"many2many:board_tags;" json:"-"`
	Name  string `gorm:"unique" json:"tag-name"`
	Count uint   `json:"-"`
}

type SourceLink struct {
	ID              uuid.UUID `gorm:"primary_key;type:uuid;default:gen_random_uuid()" json:"-"`
	RefArticleTitle string    `json:"ref-article-title-in-other-website"`
	RefArticleLink  string    `json:"ref-article-link-in-other-website"`
	BoardID         uuid.UUID `json:"-"`
}

type KeywordsWithCount struct {
	ID      uuid.UUID `gorm:"primary_key;type:uuid;default:gen_random_uuid()" json:"-"`
	Keyword string    `json:"keyword"`
	Count   int       `json:"count"`
	BoardID uuid.UUID `json:"-"`
}

type Position struct {
	ID       uuid.UUID `gorm:"primary_key;type:uuid;default:gen_random_uuid()" json:"position-id"`
	BoardID  uuid.UUID `gorm:"type:uuid" json:"-"`
	Name     string    `json:"position-name"`
	Comments []Comment `gorm:"ForeignKey:PositionID" json:"comments"`
}

type Comment struct {
	ID              uuid.UUID      `gorm:"primary_key;type:uuid;default:gen_random_uuid()" json:"comment-id"`
	PositionID      uuid.UUID      `gorm:"type:uuid" json:"post-position-id"`
	CommentedUserID string         `json:"commented-user-id"`
	CommentUserName string         `json:"commented-user-name"`
	CommentedDate   time.Time      `json:"commented-date"`
	ReportedCount   int            `json:"-"`
	Reporter        pq.StringArray `gorm:"type:text[]" json:"-"`
	Content         string         `json:"content"`
	Liker           pq.StringArray `gorm:"type:text[]" json:"-"`
	NumOfLiker      int            `json:"number-of-like"`
	IsLike          bool           `gorm:"-" json:"is-user-liked"`
	IsRePorted      bool           `gorm:"-" json:"is-user-reported"`
}

func NewBoardModel() (*DataModel, error) {
	dsn := "host=postgres user=bing password=******* dbname=board port=5432 sslmode=disable"
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, err
	}
	log.Println("connect to database!")

	db.AutoMigrate(&Board{}, &Tag{}, &SourceLink{}, &KeywordsWithCount{}, &Position{}, &Comment{})
	return &DataModel{db: db}, nil
}
func (d *DataModel) GetParticularBoard(BoardID *uuid.UUID) (*Board, error) {
	var result Board
	err := d.db.Preload("Positions.Comments").Preload(clause.Associations).Find(&result, BoardID).Error
	if err != nil {
		return nil, err
	}
	log.Println(result)
	if result.ID.String() == "00000000-0000-0000-0000-000000000000" {
		return &result, nil
	}
	for _, j := range result.Tags {
		err := d.db.Model(&j).Update("count", j.Count+1).Error
		if err != nil {
			log.Println(err)
			return nil, err
		}
	}
	err = d.db.Model(&result).Update("count", result.Count+1).Error
	if err != nil {
		log.Println(err)
		return nil, err
	}
	return &result, nil
}
func (d *DataModel) PostComment(comment *Comment) error {
	err := d.db.Create(comment).Error
	if err != nil {
		return err
	}
	return nil
}
func (d *DataModel) LikeComment(UserID *string, CommentID *uuid.UUID, isLike bool) (*Comment, error) {
	var comment Comment
	err := d.db.Where("ID = ?", CommentID).Find(&comment).Error
	if err != nil {
		log.Println(err)
		return nil, err
	}
	if isLike {
		err = d.db.Model(&comment).Update("num_of_liker", comment.NumOfLiker+1).Error
		if err != nil {
			log.Println(err)
			return nil, err
		}
		updateExpression := gorm.Expr("array_append(liker,?)", UserID)
		err = d.db.Model(&comment).Update("liker", updateExpression).Error
		if err != nil {
			log.Println(err)
			return nil, err
		}
	} else {
		err = d.db.Model(&comment).Update("num_of_liker", comment.NumOfLiker-1).Error
		if err != nil {
			log.Println(err)
			return nil, err
		}
		updateExpression := gorm.Expr("array_remove(liker,?)", UserID)
		err = d.db.Model(&comment).Update("liker", updateExpression).Error
		if err != nil {
			log.Println(err)
			return nil, err
		}
	}

	return &comment, nil
}
func (d *DataModel) CollectBoard(UserID *string, BoardID *uuid.UUID, isLike bool) error {
	var board Board
	err := d.db.Find(&board, BoardID).Error
	if err != nil {
		log.Println(err)
		return err
	}
	if isLike {
		updateExpression := gorm.Expr("array_append(collecter,?)", UserID)
		err = d.db.Model(&board).Update("collecter", updateExpression).Error
		if err != nil {
			log.Println(err)
			return err
		}
	} else {
		updateExpression := gorm.Expr("array_remove(collecter,?)", UserID)
		err = d.db.Model(&board).Update("collecter", updateExpression).Error
		if err != nil {
			log.Println(err)
			return err
		}
	}
	return nil
}
func (d *DataModel) LikeBoard(UserID *string, BoardID *uuid.UUID, isLike bool) error {
	var board Board
	err := d.db.Find(&board, BoardID).Error
	if err != nil {
		log.Println(err)
		return err
	}
	if isLike {
		err = d.db.Model(&board).Update("num_of_like", board.NumOfLike+1).Error
		if err != nil {
			log.Println(err)
			return err
		}
		updateExpression := gorm.Expr("array_append(liker,?)", UserID)
		err = d.db.Model(&board).Update("liker", updateExpression).Error
		if err != nil {
			log.Println(err)
			return err
		}
	} else {
		err = d.db.Model(&board).Update("num_of_like", board.NumOfLike-1).Error
		if err != nil {
			log.Println(err)
			return err
		}
		updateExpression := gorm.Expr("array_remove(liker,?)", UserID)
		err = d.db.Model(&board).Update("liker", updateExpression).Error
		if err != nil {
			log.Println(err)
			return err
		}
	}
	return nil
}
func (d *DataModel) CheckPersonalInfoForBoard(UserID string, Board *Board) {
	for _, t := range Board.Liker {
		if t == UserID {
			Board.IsUserLike = true
			break
		}
	}
	for _, t := range Board.Collecter {
		if t == UserID {
			Board.IsUserCollected = true
			break
		}
	}
	for i, t := range Board.Positions {
		for j, k := range t.Comments {
			for _, m := range k.Liker {
				if m == UserID {
					Board.Positions[i].Comments[j].IsLike = true
					break
				}
			}
			for _, m := range k.Reporter {
				if m == UserID {
					Board.Positions[i].Comments[j].IsRePorted = true
					break
				}
			}
		}
	}
}
func (d *DataModel) ReportComment(UserID *string, CommentID *uuid.UUID, isReported bool) error {
	var comment Comment
	err := d.db.Where("ID = ?", CommentID).Find(&comment).Error
	if err != nil {
		log.Println(err)
		return err
	}
	if isReported {
		err = d.db.Model(&comment).Update("reporte_count", comment.NumOfLiker+1).Error
		if err != nil {
			log.Println(err)
			return err
		}
		updateExpression := gorm.Expr("array_append(reporter,?)", UserID)
		err = d.db.Model(&comment).Update("reporter", updateExpression).Error
		if err != nil {
			log.Println(err)
			return err
		}
	} else {
		err = d.db.Model(&comment).Update("reporte_count", comment.NumOfLiker-1).Error
		if err != nil {
			log.Println(err)
			return err
		}
		updateExpression := gorm.Expr("array_remove(reporter,?)", UserID)
		err = d.db.Model(&comment).Update("reporter", updateExpression).Error
		if err != nil {
			log.Println(err)
			return err
		}
	}
	return nil
}
func (d *DataModel) GetPopularTags() (*[]Tag, error) {
	var tags []Tag
	err := d.db.Model(&Tag{}).Order("count desc").Find(&tags).Error
	if err != nil {
		return nil, err
	}
	return &tags, nil
}
func (d *DataModel) GetPopularBoards() (*[]Board, error) {
	var result []Board
	err := d.db.Preload("Tags").Preload("KeywordsWithCounts").Find(&result).Error
	if err != nil {
		return nil, err
	}
	return &result, nil
}

func (d *DataModel) SearchBoardByKeywordsOrTags(list []string) (*[]Board, error) {
	var result []Board
	query := d.db.Preload("Tags").Preload("KeywordsWithCounts").Where("id IN (?)", d.db.Table("board_tags").Select("board_id").Where("tag_id in (?)", d.db.Table("tags").Select("id").Where("name in ?", list)))
	for _, j := range list {
		query = query.Or("title like ?", "%"+j+"%")
	}
	err := query.Find(&result).Error
	if err != nil {
		return nil, err
	}
	return &result, nil
}
func (d *DataModel) SetFakeValue() error {
	data1 := &Board{
		Date:               time.Now(),
		Title:              "婚前同居",
		ImageCloud:         "localhost/board/get-photo/testboard.png",
		SourceLinks:        []SourceLink{{RefArticleTitle: "helloPtt", RefArticleLink: "ptt://"}, {RefArticleTitle: "helloDcard", RefArticleLink: "dcard://"}},
		Tags:               []Tag{{Name: "自由", Count: 0}, {Name: "道德", Count: 0}},
		KeywordsWithCounts: []KeywordsWithCount{{Keyword: "懷孕", Count: 20}, {Keyword: "婚姻", Count: 50}},
		Positions:          []Position{{Name: "yes", Comments: []Comment{}}, {Name: "no", Comments: []Comment{}}},
		NumOfLike:          0,
		Liker:              []string{},
		Collecter:          []string{},
		Count:              0,
	}
	for i, t := range data1.Tags {
		var existingTag Tag
		if err := d.db.Where("name = ?", t.Name).First(&existingTag).Error; err == nil {
			data1.Tags[i] = existingTag
		} else {
			continue
		}
	}
	err := d.db.Create(data1).Error
	if err != nil {
		log.Println(err)
		log.Println("insert error")
	}
	data2 := &Board{
		Date:               time.Now(),
		Title:              "火龍果害我以為得大腸癌==",
		ImageCloud:         "hello",
		SourceLinks:        []SourceLink{{RefArticleTitle: "helloPtt", RefArticleLink: "ptt://"}, {RefArticleTitle: "helloDcard", RefArticleLink: "dcard://"}},
		Tags:               []Tag{{Name: "水果", Count: 0}, {Name: "癌症", Count: 0}},
		KeywordsWithCounts: []KeywordsWithCount{{Keyword: "笑死", Count: 20}, {Keyword: "紅紅ㄉ", Count: 20}},
		Positions:          []Position{{Name: "yes", Comments: []Comment{}}, {Name: "no", Comments: []Comment{}}},
		NumOfLike:          0,
		Liker:              []string{},
		Collecter:          []string{},
		Count:              0,
	}
	for i, t := range data2.Tags {
		var existingTag Tag
		if err := d.db.Where("name = ?", t.Name).First(&existingTag).Error; err == nil {
			data2.Tags[i] = existingTag
		} else {
			continue
		}
	}
	err = d.db.Create(data2).Error
	if err != nil {
		log.Println(err)
		log.Println("insert error")
	}

	return nil
}
