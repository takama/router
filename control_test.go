package router

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

var params = []Param{
	{"name", "John"},
	{"age", "32"},
	{"gender", "M"},
}

var testParamsData = `[
  {
    "key": "name",
    "value": "John"
  },
  {
    "key": "age",
    "value": "32"
  },
  {
    "key": "gender",
    "value": "M"
  }
]`

func TestControlParamsSetGet(t *testing.T) {

	c := new(Control)
	c.Set(params...)
	for _, param := range params {
		if c.Get(param.Key) != param.Value {
			t.Error("Expected for", param.Key, ":", param.Value, ", got", c.Get(param.Key))
		}
	}
}

func TestControlCode(t *testing.T) {
	c := new(Control)
	// code transcends, must be less than 600
	c.Code(606)
	if c.code != 0 {
		t.Error("Expected code", "0", "got", c.code)
	}
	c.Code(404)
	if c.code != 404 {
		t.Error("Expected code", "404", "got", c.code)
	}
}

func TestMetaData(t *testing.T) {
	req, err := http.NewRequest("GET", "module/:data", nil)
	if err != nil {
		t.Error(err)
	}
	hd := Header{
		APIVersion: "2.1",
	}
	c := new(Control)
	c.CompactJSON(true).UseMetaData()
	trw := httptest.NewRecorder()
	c.Writer, c.Request = trw, req
	c.APIVersion(hd.APIVersion).HeaderContext(hd.Context).Body(nil)
	if content, err := json.Marshal(hd); err == nil {
		if trw.Body.String() != string(content) {
			t.Error("Expected", string(content), "got", trw.Body.String())
		}
	} else {
		t.Error(err)
	}
	trw = httptest.NewRecorder()
	c.Writer = trw
	hd.Context = "bart"
	hd.Method = "people.get"
	hd.ID = "id17"
	params := []Param{{Key: "userId", Value: "@me"}, {Key: "groupId", Value: "@self"}}
	hd.Params = params
	c.Set(params...).HeaderContext(hd.Context).Method(hd.Method).ID(hd.ID).Body(nil)
	if content, err := json.Marshal(hd); err == nil {
		if trw.Body.String() != string(content) {
			t.Error("Expected", string(content), "got", trw.Body.String())
		}
	} else {
		t.Error(err)
	}
	trw = httptest.NewRecorder()
	c.Writer = trw
	altParams := map[string]string{"userId": "@me", "groupId": "@self"}
	hd.Params = altParams
	c.SetParams(altParams).Body(nil)
	if content, err := json.Marshal(hd); err == nil {
		if trw.Body.String() != string(content) {
			t.Error("Expected", string(content), "got", trw.Body.String())
		}
	} else {
		t.Error(err)
	}
}

func TestErrorData(t *testing.T) {
	req, err := http.NewRequest("GET", "module/:data", nil)
	if err != nil {
		t.Error(err)
	}
	ed := ErrorHeader{
		Code:    http.StatusBadRequest,
		Message: "Unexpected parameter :data",
	}
	c := new(Control)
	c.CompactJSON(true)
	trw := httptest.NewRecorder()
	c.Writer, c.Request = trw, req
	c.SetError(ed.Code, ed.Message).Body(nil)
	if content, err := json.Marshal(map[string]ErrorHeader{"error": ed}); err == nil {
		if trw.Body.String() != string(content) {
			t.Error("Expected", string(content), "got", trw.Body.String())
		}
	} else {
		t.Error(err)
	}
	trw = httptest.NewRecorder()
	c.Writer = trw
	errors := []Error{
		{Message: "File Not Found"},
	}
	ed.Errors = errors
	c.AddError(errors...).Body(nil)
	if content, err := json.Marshal(map[string]ErrorHeader{"error": ed}); err == nil {
		if trw.Body.String() != string(content) {
			t.Error("Expected", string(content), "got", trw.Body.String())
		}
	} else {
		t.Error(err)
	}
}

func TestControlBody(t *testing.T) {
	req, err := http.NewRequest("GET", "hello/:name", nil)
	if err != nil {
		t.Error(err)
	}
	c := new(Control)
	trw := httptest.NewRecorder()
	c.Writer, c.Request = trw, req
	c.Body("Hello")
	if trw.Body.String() != "Hello" {
		t.Error("Expected", "Hello", "got", trw.Body.String())
	}
	trw = httptest.NewRecorder()
	c.Writer = trw
	c.Body(123)
	if trw.Body.String() != "123" {
		t.Error("Expected", "123", "got", trw.Body.String())
	}
	trw = httptest.NewRecorder()
	c.Writer = trw
	c.Body(123.1)
	if trw.Body.String() != "123.1" {
		t.Error("Expected", "123.1", "got", trw.Body.String())
	}
	trw = httptest.NewRecorder()
	c.Writer = trw
	c.Body(params)
	if trw.Body.String() != testParamsData {
		t.Error("Expected", testParamsData, "got", trw.Body.String())
	}
}
