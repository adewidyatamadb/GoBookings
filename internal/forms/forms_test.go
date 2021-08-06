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
	r := httptest.NewRequest("POST", "/whatever", nil)
	form := New(r.PostForm)

	if form.Has("a") {
		t.Error("form shows has field when it does not")
	}

	postedData := url.Values{}
	postedData.Add("a", "a")

	r, _ = http.NewRequest("POST", "/whatever", nil)
	r.PostForm = postedData
	form = New(r.PostForm)
	if !form.Has("a") {
		t.Error("form shows it does not have field when it does")
	}
}

func TestForm_MinLength(t *testing.T) {
	invalidData := url.Values{}
	invalidData.Add("a", "a")
	r, _ := http.NewRequest("POST", "/whatever", nil)
	r.PostForm = invalidData
	form := New(r.PostForm)

	form.MinLength("a", 5)
	if form.Valid() {
		t.Error("form shows valid input length when it does not")
	}

	validData := url.Values{}
	validData.Add("a", "valid")
	r, _ = http.NewRequest("POST", "/whatever", nil)
	r.PostForm = validData
	form = New(r.PostForm)

	form.MinLength("a", 5)
	if !form.Valid() {
		t.Error("form shows invalid input length when it does")
	}

}

func TestForm_IsEmail(t *testing.T) {
	invalidData := url.Values{}
	invalidData.Add("a", "invalid")
	r, _ := http.NewRequest("POST", "/whatever", nil)
	r.PostForm = invalidData
	form := New(r.PostForm)

	form.IsEmail("a")
	if form.Valid() {
		t.Error("form shows valid email address when it does not")
	}

	validData := url.Values{}
	validData.Add("a", "valid@email.com")
	r, _ = http.NewRequest("POST", "/whatever", nil)
	r.PostForm = validData
	form = New(r.PostForm)

	form.IsEmail("a")
	if !form.Valid() {
		t.Error("form shows invalid email address when it does")
	}
}
