package forms

import (
	"net/http"
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
	r := httptest.NewRequest("POST", "/whatever", nil)
	form := New(r.PostForm)

	form.Required("a", "b", "c")
	if form.Valid() {
		t.Error("form shows valid when required fields missing")
	}

	postedData := url.Values{}
	postedData.Add("a", "a")
	postedData.Add("b", "a")
	postedData.Add("c", "a")

	r, _ = http.NewRequest("POST", "/whatever", nil)
	r.PostForm = postedData
	form = New(r.PostForm)
	form.Required("a", "b", "c")
	for !form.Valid() {
		t.Error("form shows does not have required fields when it does")
	}

}

func TestForm_Has(t *testing.T) {
	postedData := url.Values{}
	form := New(postedData)

	if form.Has("a") {
		t.Error("form shows has field when it does not")
	}

	postedData = url.Values{}
	postedData.Add("a", "a")
	form = New(postedData)
	if !form.Has("a") {
		t.Error("form shows it does not have field when it does")
	}
}

func TestForm_MinLength(t *testing.T) {
	invalidData := url.Values{}
	invalidData.Add("a", "a")
	form := New(invalidData)

	form.MinLength("a", 5)
	if form.Valid() {
		t.Error("form shows valid input length when it does not")
	}

	isError := form.Errors.Get("a")
	if isError == "" {
		t.Error("should have an error, but did not get one")
	}

	form.MinLength("non-existent", 5)
	if form.Valid() {
		t.Error("form shows valid input length when the field does not exist")
	}

	validData := url.Values{}
	validData.Add("valid", "valid")
	form = New(validData)

	form.MinLength("valid", 5)
	if !form.Valid() {
		t.Error("form shows invalid input length when it does")
	}

	isError = form.Errors.Get("valid")
	if isError != "" {
		t.Error("should not get an error but got one")
	}

}

func TestForm_IsEmail(t *testing.T) {
	invalidData := url.Values{}
	invalidData.Add("a", "invalid")
	form := New(invalidData)

	form.IsEmail("a")
	if form.Valid() {
		t.Error("form shows valid email address when it does not")
	}

	form.IsEmail("non-existent")
	if form.Valid() {
		t.Error("form shows valid email address then the field does not exist")
	}

	validData := url.Values{}
	validData.Add("a", "valid@email.com")
	form = New(validData)

	form.IsEmail("a")
	if !form.Valid() {
		t.Error("form shows invalid email address when it does")
	}
}
