package main

import (
	//"fmt"
	"testing"
	"strconv"
)


func Test_deleteTask(t *testing.T) {
	inputs := []struct {
		id   int
		result string
	}{
		{id: 12, result: "1"},
		{id: 13, result: "1"},
		{id: 14, result: "1"},
		{id: 16, result: "1"},
		{id: 17, result: "1"},
		{id: 18, result: "1"},
		{id: 19, result: "1"},
		{id: 555, result: "0"},
		{id: 1500, result: "0"},		
		{id: -15, result: "0"},		
		{id: 190, result: "0"},
		{id: 99, result: "0"},
		{id: 2220, result: "0"},
		{id: -22, result: "0"},		
	}

	for _, item := range inputs {

		resultInt, _ := deleteTask(item.id)
		result := strconv.Itoa(int(resultInt))

		if result != item.result {
			t.Errorf("\"deleteTask('%d')\" failed, expected -> %v, result -> %v", item.id, item.result, result)
		} else {
			t.Logf("\"deleteTask('%d')\" succeded, expected -> %v, result -> %v", item.id, item.result, result)
		}
	}
}


