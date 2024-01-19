package pgvector_test

import (
	"math"
	"reflect"
	"testing"

	"github.com/pgvector/pgvector-go"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type GormItem struct {
	gorm.Model
	Embedding pgvector.Vector `gorm:"type:vector(3)"`
}

func CreateGormItems(db *gorm.DB) {
	items := []GormItem{
		GormItem{Embedding: pgvector.NewVector([]float32{1, 1, 1})},
		GormItem{Embedding: pgvector.NewVector([]float32{2, 2, 2})},
		GormItem{Embedding: pgvector.NewVector([]float32{1, 1, 2})},
	}

	result := db.Create(items)

	if result.Error != nil {
		panic(result.Error)
	}
}

func TestGorm(t *testing.T) {
	db, err := gorm.Open(postgres.Open("dbname=pgvector_go_test"), &gorm.Config{})
	if err != nil {
		panic(err)
	}

	db.Exec("CREATE EXTENSION IF NOT EXISTS vector")
	db.Exec("DROP TABLE IF EXISTS gorm_items")

	db.AutoMigrate(&GormItem{})

	CreateGormItems(db)

	var items []GormItem
	db.Clauses(clause.OrderBy{
		Expression: clause.Expr{SQL: "embedding <-> ?", Vars: []interface{}{pgvector.NewVector([]float32{1, 1, 1})}},
	}).Limit(5).Find(&items)
	if items[0].ID != 1 || items[1].ID != 3 || items[2].ID != 2 {
		t.Errorf("Bad ids")
	}
	if !reflect.DeepEqual(items[1].Embedding.Slice(), []float32{1, 1, 2}) {
		t.Errorf("Bad embedding")
	}

	var distances []float64
	db.Model(&GormItem{}).Select("embedding <-> ?", pgvector.NewVector([]float32{1, 1, 1})).Order("id").Find(&distances)
	if distances[0] != 0 || distances[1] != math.Sqrt(3) || distances[2] != 1 {
		t.Errorf("Bad distances")
	}
}
