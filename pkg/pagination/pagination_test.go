package pagination_test

import (
	"fmt"
	"testing"

	"github.com/chan-jui-huang/go-backend-package/pkg/pagination"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type User struct {
	ID   uint
	Name string
	Age  int
}

func newTestDB(t *testing.T) *gorm.DB {
	dsn := fmt.Sprintf("file:%s?mode=memory&cache=shared", t.Name())
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to open db: %v", err)
	}
	if err := db.AutoMigrate(&User{}); err != nil {
		t.Fatalf("failed migrate: %v", err)
	}
	return db
}

func seedUsers(t *testing.T, db *gorm.DB, n int) {
	for i := 1; i <= n; i++ {
		u := User{Name: fmt.Sprintf("user%03d", i), Age: i}
		if err := db.Create(&u).Error; err != nil {
			t.Fatalf("failed to seed user: %v", err)
		}
	}
}

func TestGetTotalAndLastPage(t *testing.T) {
	db := newTestDB(t)
	seedUsers(t, db, 25)

	p := pagination.NewPaginator(db.Model(&User{}), nil, nil, 1, 10)
	total, last := p.GetTotalAndLastPage()
	if total != 25 {
		t.Fatalf("expected total 25, got %d", total)
	}
	if last != 3 {
		t.Fatalf("expected last page 3, got %d", last)
	}
}

func TestExecutePagination(t *testing.T) {
	db := newTestDB(t)
	seedUsers(t, db, 15)

	p := pagination.NewPaginator(db.Model(&User{}), nil, nil, 2, 10) // page 2, perPage 10 -> should return 5 records
	var users []User
	res := p.Execute(&users)
	if res.Error != nil {
		t.Fatalf("execute error: %v", res.Error)
	}
	if len(users) != 5 {
		t.Fatalf("expected 5 users on page 2, got %d", len(users))
	}
	// ensure offset applied by checking first returned user's Age is 11
	if users[0].Age != 11 {
		t.Fatalf("expected first user age 11 on page 2, got %d", users[0].Age)
	}
}

func TestAddWhereConditionsCallsFunctions(t *testing.T) {
	db := newTestDB(t)
	seedUsers(t, db, 10)

	called := []any{}
	whereMap := pagination.WhereConditionMap{
		"age": func(db *gorm.DB, value any) {
			// record that the function was invoked with the value
			called = append(called, value)
		},
	}

	p := pagination.NewPaginator(db.Model(&User{}), whereMap, nil, 1, 10)
	p.AddWhereConditions(map[string]any{"age": 5})

	if len(called) != 1 {
		t.Fatalf("expected where function to be called once, called=%d", len(called))
	}
	if called[0] != 5 {
		t.Fatalf("expected where function to be called with 5, got %v", called[0])
	}
}

func TestOrderByNoPanic(t *testing.T) {
	db := newTestDB(t)
	seedUsers(t, db, 5)

	orders := map[string]string{"name_desc": "name desc"}
	p := pagination.NewPaginator(db.Model(&User{}), nil, orders, 1, 10)
	// should not panic or error even if OrderBy doesn't mutate internal state
	p.OrderBy("name_desc")
	var users []User
	res := p.Execute(&users)
	if res.Error != nil {
		t.Fatalf("execute error after order: %v", res.Error)
	}
	if len(users) != 5 {
		t.Fatalf("expected 5 users, got %d", len(users))
	}
}
