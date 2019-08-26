package main

import "testing"
import "fmt"
import "time"

func TestCelebrateBirthday(t *testing.T) {
	tables := []struct {
		id int
		birthday string
		isBD bool
	}{
		{1, "1999", true},
		{2, "1989", true},
		{3, "1999-06-21", false},
		{4, "1980", true},
		{5, "1992-03-02", false},
	}

	c, err := createClient()
	if err != nil {
		t.Errorf("Error setting redis data: %v", err.Error())
	}
	c.rc.FlushAll()
	d := time.Now()
	for _, table := range tables {
		key := fmt.Sprintf(APP_PREFIX+"user-%d", table.id)
		if table.isBD {
			table.birthday = table.birthday + fmt.Sprintf("-%02d-%02d", d.Month(), d.Day())
		}
		c.rc.HSet(key, "birthday", table.birthday)
	}
	c.rc.Set(APP_PREFIX+"user-ids", tables[len(tables)-1].id, 0)

	for i:=0; i<tables[len(tables)-1].id; i++ {
		user := Find(c, tables[i].id)
		if user != nil {
			b := user.isBirthday()
			if b != tables[i].isBD {
				t.Errorf("%d Birthday: got: %v, want: %v.", tables[i].id, b, tables[i].isBD)
			}
		}
	}
}

