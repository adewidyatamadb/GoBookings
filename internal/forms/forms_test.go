package forms

import (
	"net/http/httptest"
	"net/url"
	"testing"
)

func TestForm_Valid(t *testing.T) {
	r := httptest.NewRequest("POST", "/whatever", nil)
	form := New(r.PostForm)

	isValid := form.Valid()
	if !isValid {
		t.Error("got invalid when should have been valid")
	}
}

func TestForm_Required(t *testing.T) {
	var tableTest = []struct {
		name  string
		valid bool
	}{
		{"missing field", false},
		{"valid form", true},
	}

	for _, test := range tableTest {
		r := httptest.NewRequest("POST", "/whatever", nil)
		var f *Form
		postedData := url.Values{}
		if test.valid {
			postedData.Add("a", "a")
			postedData.Add("b", "a")
			postedData.Add("c", "a")
			r.PostForm = postedData
			form := New(r.PostForm)
			f = form
		} else {
			form := New(r.PostForm)
			f = form
		}

		f.Required("a", "b", "c")
		if f.Valid() != test.valid {
			t.Errorf("case - %s: test returned %v wanted %v", test.name, f.Valid(), test.valid)
		}
	}
}

func TestForm_Has(t *testing.T) {
	var tableTest = []struct {
		name     string
		hasField bool
	}{
		{"field not exist", false},
		{"field exist", true},
	}

	for _, test := range tableTest {
		postedData := url.Values{}

		if test.hasField {
			postedData.Add("a", "a")
		}
		form := New(postedData)

		if form.Has("a") != test.hasField {
			t.Errorf("case - %s: test return %v, wanted %v", test.name, form.Has("a"), test.hasField)
		}
	}
}

func TestForm_MinLength(t *testing.T) {
	var tableTest = []struct {
		name       string
		param      string
		valid      bool
		fieldExist bool
		err        string
	}{
		{"invalid data", "a", false, true, "This field must be at least 5 characters long"},
		{"non-existent field", "a", false, false, ""},
		{"valid", "valid", true, true, ""},
	}

	for _, test := range tableTest {
		data := url.Values{}
		data.Add("field", test.param)
		form := New(data)
		if test.fieldExist {
			form.MinLength("field", 5)
		} else {
			form.MinLength("not-exist", 5)
		}

		if form.Valid() != test.valid {
			t.Errorf("case - %s: test return %v, wanted %v", test.name, form.Valid(), test.valid)
		}

		err := form.Errors.Get("field")
		if err != test.err {
			t.Errorf("case - %s: test return %s, wanted %s", test.name, err, test.err)
		}
	}

}

func TestForm_IsEmail(t *testing.T) {
	var tableTest = []struct {
		name       string
		fieldName  string
		value      string
		fieldExist bool
		valid      bool
	}{
		{"invalid data", "field", "invalid", true, false},
		{"invalid data", "field", "invalid", false, false},
		{"valid data", "field", "valid@email.com", true, true},
	}

	for _, test := range tableTest {
		data := url.Values{}
		data.Add(test.fieldName, test.value)
		form := New(data)
		if test.fieldExist {
			form.IsEmail(test.fieldName)
		} else {
			form.IsEmail("non-existent-field")
		}

		if form.Valid() != test.valid {
			t.Errorf("case - %s: test returned %v, wanted %v", test.name, form.Valid(), test.valid)
		}
	}
}
