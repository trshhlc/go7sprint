package main

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCafeNegative(t *testing.T) {
	handler := http.HandlerFunc(mainHandle)

	requests := []struct {
		request string
		status  int
		message string
	}{
		{"/cafe", http.StatusBadRequest, "unknown city"},
		{"/cafe?city=omsk", http.StatusBadRequest, "unknown city"},
		{"/cafe?city=tula&count=na", http.StatusBadRequest, "incorrect count"},
	}
	for _, v := range requests {
		response := httptest.NewRecorder()
		req := httptest.NewRequest("GET", v.request, nil)
		handler.ServeHTTP(response, req)

		assert.Equal(t, v.status, response.Code)
		assert.Equal(t, v.message, strings.TrimSpace(response.Body.String()))
	}
}

func TestCafeWhenOk(t *testing.T) {
	handler := http.HandlerFunc(mainHandle)

	requests := []string{
		"/cafe?count=2&city=moscow",
		"/cafe?city=tula",
		"/cafe?city=moscow&search=ложка",
	}
	for _, v := range requests {
		response := httptest.NewRecorder()
		req := httptest.NewRequest("GET", v, nil)

		handler.ServeHTTP(response, req)

		assert.Equal(t, http.StatusOK, response.Code)
	}
}

const testCity = "moscow"

func TestCafeCount(t *testing.T) {
	// проверяет работу сервера при разных значениях параметра count
	handler := http.HandlerFunc(mainHandle)

	requests := []struct {
		count int // передаваемое значение count
		want  int // ожидаемое количество кафе в ответе
	}{
		{count: 0, want: 0},
		{count: 1, want: 1},
		{count: 2, want: 2},
		{count: 100, want: min(len(cafeList[testCity]), 100)},
	}

	for _, v := range requests {
		url := fmt.Sprintf("/cafe?city=%s&count=%d", testCity, v.count)
		response := httptest.NewRecorder()
		req := httptest.NewRequest("GET", url, nil)

		handler.ServeHTTP(response, req)

		r := strings.TrimSpace(response.Body.String())
		var count int
		if r != "" {
			count = len(strings.Split(r, ","))
		}

		require.Equal(t, http.StatusOK, response.Code)
		assert.Equal(t, v.want, count)
	}
}

func TestCafeSearch(t *testing.T) {
	// проверяет результат поиска кафе по указанной подстроке в параметре search
	handler := http.HandlerFunc(mainHandle)

	requests := []struct {
		search    string // передаваемое значение search
		wantCount int    // ожидаемое количество кафе в ответе
	}{
		{search: "search=фасоль", wantCount: 0},
		{search: "search=кофе", wantCount: 2},
		{search: "search=вилка", wantCount: 1},
	}

	for _, v := range requests {
		url := fmt.Sprintf("/cafe?city=%s&%s", testCity, v.search)
		response := httptest.NewRecorder()
		req := httptest.NewRequest("GET", url, nil)

		handler.ServeHTTP(response, req)

		responseString := strings.TrimSpace(response.Body.String())
		var count int
		if responseString != "" {
			responseArr := strings.Split(responseString, ",")
			count = len(responseArr)
			for _, cafe := range responseArr {
				search := strings.Split(v.search, "=")[1]
				assert.True(t, strings.Contains(strings.ToLower(cafe), strings.ToLower(search)))
			}
		}

		require.Equal(t, http.StatusOK, response.Code)
		assert.Equal(t, v.wantCount, count)
	}
}
