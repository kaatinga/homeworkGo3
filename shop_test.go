package models

import (
	"testing"
)

func TestShop_AddGood(t *testing.T) {

	s := NewShop()

	type args struct {
		name  string
		unit  string
		price uint64
	}

	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{"ok", args{
			name:  "Яблоко",
			unit:  "Кг.",
			price: 100,
		}, false},
		{"!ok1", args{
			name:  "",
			unit:  "Кг.",
			price: 100,
		}, true},
		{"!ok2", args{
			name:  "Груша",
			unit:  "Кг.",
			price: 0,
		}, true},
	}

	var oldEntryID uint16

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			oldEntryID = s.goods.entryID

			if err := s.AddGood(tt.args.name, tt.args.unit, tt.args.price); (err != nil) != tt.wantErr {
				t.Errorf("AddGood() error = %v, wantErr %v", err, tt.wantErr)
			}

			if !tt.wantErr && (s.goods.entryID != oldEntryID+1) {
				t.Errorf("entryID = %v, but must be %v", s.goods.entryID, oldEntryID+1)
			}

			if tt.wantErr && (s.goods.entryID != oldEntryID) {
				t.Errorf("entryID = %v, but must be %v", s.goods.entryID, oldEntryID)
			}
		})
	}
}
