package main

import (
	//"fmt"
	"testing"
	"strings"
	"io/ioutil"
	"strconv"
	"net/http"
    "net/http/httptest"
)



func Test_deleteHandler(t *testing.T) {
	inputs := []struct {
		id   int
		expected string
	}{
		{id: 1, expected: "Id doesn't exist"},
		{id: 44, expected: "successfully deleted!"},
		{id: 145, expected: "Id doesn't exist"},
		{id: 45, expected: "successfully deleted!"},
		{id: -5, expected: "Id must be a valid number"},
		{id: 46, expected: "successfully deleted!"},
		{id: 47, expected: "successfully deleted!"},
		{id: 555, expected: "Id doesn't exist"},			
	}

	for _, item := range inputs {

		url := "/deleteTask?id="+strconv.Itoa(item.id)
		req := httptest.NewRequest(http.MethodGet, url, nil)
		w := httptest.NewRecorder()
		deleteHandler(w, req)
		res := w.Result()
		defer res.Body.Close()
		data, err := ioutil.ReadAll(res.Body)
		if err != nil {
			t.Errorf("expected error to be nil got %v", err)
		}
		result := strings.TrimSpace(string(data))
		
		if result != item.expected {
			t.Errorf("\"/deleteTask?id=%d\" failed, expected -> %v, result -> %v", item.id, item.expected, result)
		} else {
			t.Logf("\"/deleteTask?id=%d\" succeded, expected -> %v, result -> %v", item.id, item.expected, result)
		}
	}
}


