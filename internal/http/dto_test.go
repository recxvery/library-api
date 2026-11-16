package http

import (
	"errors"
	"reflect"
	"strings"
	"testing"

	"rest-api-app/internal/core"
	"rest-api-app/internal/db"
)

func TestValidateToCreate(t *testing.T) {
	validDTO := BookDTO{
		Author:      "George Orwell",
		Title:       "1984",
		Description: "Dystopian novel",
		Pages:       328,
		Year:        1949,
	}

	validModel := db.BookModel{
		Author:      "George Orwell",
		Title:       "1984",
		Description: Ptr("Dystopian novel"),
		PagesCount:  328,
		Year:        1949,
	}

	tests := []struct {
		name    string
		input   BookDTO
		want    db.BookModel
		wantErr bool
	}{
		{
			name:  "valid book",
			input: validDTO,
			want:  validModel,
		},
		{
			name: "empty description is allowed",
			input: BookDTO{
				Author: "Frank Herbert",
				Title:  "Dune",
				Pages:  412,
				Year:   1965,
			},
			want: db.BookModel{
				Author:      "Frank Herbert",
				Title:       "Dune",
				Description: Ptr(""),
				PagesCount:  412,
				Year:        1965,
			},
		},
		{
			name: "page count above the minimum",
			input: BookDTO{
				Author: "Antoine de Saint-Exupery",
				Title:  "The Little Prince",
				Pages:  31,
				Year:   1943,
			},
			want: db.BookModel{
				Author:      "Antoine de Saint-Exupery",
				Title:       "The Little Prince",
				Description: Ptr(""),
				PagesCount:  31,
				Year:        1943,
			},
		},
		{
			name: "year above the minimum",
			input: BookDTO{
				Author: "Unknown",
				Title:  "Ancient manuscript",
				Pages:  100,
				Year:   101,
			},
			want: db.BookModel{
				Author:      "Unknown",
				Title:       "Ancient manuscript",
				Description: Ptr(""),
				PagesCount:  100,
				Year:        101,
			},
		},
		{
			name: "empty author",
			input: BookDTO{
				Title:       "1984",
				Description: "Dystopian novel",
				Pages:       328,
				Year:        1949,
			},
			wantErr: true,
		},
		{
			name: "empty title",
			input: BookDTO{
				Author:      "George Orwell",
				Description: "Dystopian novel",
				Pages:       328,
				Year:        1949,
			},
			wantErr: true,
		},
		{
			name: "too few pages",
			input: BookDTO{
				Author: "Raymond Smullyan",
				Title:  "Alice in Puzzle-Land",
				Pages:  30,
				Year:   1982,
			},
			wantErr: true,
		},
		{
			name: "year too small",
			input: BookDTO{
				Author: "Unknown Etruscan master",
				Title:  "Etruscan Gold Book",
				Pages:  6,
				Year:   100,
			},
			wantErr: true,
		},
		{
			name: "negative year",
			input: BookDTO{
				Author: "Unknown Etruscan master",
				Title:  "Etruscan Gold Book",
				Pages:  100,
				Year:   -660,
			},
			wantErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := ValidateToCreate(tc.input)

			if tc.wantErr {
				if !errors.Is(err, core.ErrorNotEnoughData) {
					t.Fatalf("error: got %v, want %v", err, core.ErrorNotEnoughData)
				}
				if !reflect.DeepEqual(got, db.BookModel{}) {
					t.Errorf("book: got %+v, want zero value", got)
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if !reflect.DeepEqual(got, tc.want) {
				t.Errorf("book mismatch:\ngot:  %+v\nwant: %+v", got, tc.want)
			}
		})
	}
}

func TestErrToJSON(t *testing.T) {
	data := errToJSON(core.ErrorNotEnoughData)

	if len(data) == 0 {
		t.Fatal("expected non-empty json")
	}

	got := string(data)
	if !strings.Contains(got, core.ErrorNotEnoughData.Error()) {
		t.Errorf("json does not contain error message: %s", got)
	}
}

func Ptr[T any](v T) *T {
	return &v
}
