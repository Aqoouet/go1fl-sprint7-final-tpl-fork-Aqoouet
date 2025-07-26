package main

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func getResponse(reqString string) responseData {

	handler := http.HandlerFunc(mainHandle)

	resp := httptest.NewRecorder()
	req := httptest.NewRequest("GET", reqString, nil)

	handler.ServeHTTP(resp, req)

	if resp.Code == http.StatusOK {
		if resp.Body.String() == "" {
			return responseData{resp.Code, 0, nil, resp.Body.String()}
		} else {
			slice := strings.Split(resp.Body.String(), ",")

			for i := range slice {
				slice[i] = strings.TrimSpace(slice[i])
			}

			return responseData{resp.Code, len(slice), slice, resp.Body.String()}
		}
	}
	return responseData{resp.Code, 0, nil, resp.Body.String()}
}

func buildRequest(s map[string]string) string {

	if s == nil {
		return "/cafe"
	}

	values := url.Values{}

	for k, v := range s {
		values.Set(k, v)
	}

	query := values.Encode()

	return "/cafe?" + query

}

type responseData struct {
	status int
	count  int
	slice  []string
	answer string
}

type testItem struct {
	label      string
	reqParams  map[string]string
	wantStatus int
	wantCount  int
}

const ErrMsg = "query parameters: %v\nserver response: %q\n"

func TestCafeNegative(t *testing.T) {

	tests := []testItem{
		{
			label:      "нe указано никаких параметров",
			reqParams:  nil,
			wantStatus: http.StatusBadRequest,
		},
		{
			label:      "запрошен город Омск, по которому нет данных",
			reqParams:  map[string]string{"city": "Omsk"},
			wantStatus: http.StatusBadRequest,
		},
		{
			label:      "указано некорректное значение параметра count для Тулы",
			reqParams:  map[string]string{"city": "tula", "count": "na"},
			wantStatus: http.StatusBadRequest,
		},
	}

	for _, v := range tests {
		t.Run(v.label, func(t *testing.T) {

			reqString := buildRequest(v.reqParams)
			resp := getResponse(reqString)
			require.Equal(t, v.wantStatus, resp.status, ErrMsg, v.reqParams, resp.answer)
		})
	}

}

func TestCafeWhenOk(t *testing.T) {

	tests := []testItem{
		{
			label:      "2 кафе для Москвы",
			reqParams:  map[string]string{"city": "moscow", "count": "2"},
			wantStatus: http.StatusOK,
		},
		{
			label:      "ищем кафе с 'ложка' в Москве",
			reqParams:  map[string]string{"city": "moscow", "search": "ложка"},
			wantStatus: http.StatusOK,
		},
		{
			label:      "в запросе указан только город Тула",
			reqParams:  map[string]string{"city": "tula"},
			wantStatus: http.StatusOK,
		},
	}

	for _, v := range tests {
		t.Run(v.label, func(t *testing.T) {

			reqString := buildRequest(v.reqParams)
			resp := getResponse(reqString)
			require.Equal(t, v.wantStatus, resp.status, ErrMsg, v.reqParams, resp.answer)

		})
	}

}

func TestCafeSearch(t *testing.T) {

	tests := []testItem{

		{
			label:      "слово, которое не встречается в ресторанах Москвы",
			reqParams:  map[string]string{"city": "moscow", "search": "фасоль"},
			wantStatus: http.StatusOK,
			wantCount:  0,
		},
		{
			label:      "слово, которое встречается 2 раза для Москвы",
			reqParams:  map[string]string{"city": "moscow", "search": "кофе"},
			wantStatus: http.StatusOK,
			wantCount:  2,
		},
		{
			label:      "слово, которое встречается 1 раз для Москвы",
			reqParams:  map[string]string{"city": "moscow", "search": "вилка"},
			wantStatus: http.StatusOK,
			wantCount:  1,
		},
	}

	for _, v := range tests {
		t.Run(v.label, func(t *testing.T) {

			reqString := buildRequest(v.reqParams)
			resp := getResponse(reqString)
			require.Equal(t, v.wantStatus, resp.status, ErrMsg, v.reqParams, resp.answer)

			require.Equal(t, v.wantCount, resp.count, ErrMsg, v.reqParams, resp.answer)

			if searchWord, ok := v.reqParams["search"]; ok {
			
				searchLower := strings.ToLower(searchWord)

				for _, cafe := range resp.slice {
					cafeLower := strings.ToLower(cafe)
					require.Contains(t, cafeLower, searchLower, ErrMsg, v.reqParams, resp.answer)
				}
			}
		})
	}
}

func TestCafeCount(t *testing.T) {

	tests := []testItem{

		{
			label:      "0 кафе для Москвы",
			reqParams:  map[string]string{"city": "moscow", "count": "0"},
			wantStatus: http.StatusOK,
			wantCount:  0,
		},
		{
			label:      "1 кафе для Москвы",
			reqParams:  map[string]string{"city": "moscow", "count": "1"},
			wantStatus: http.StatusOK,
			wantCount:  1,
		},
		{
			label:      "2 кафе для Москвы",
			reqParams:  map[string]string{"city": "moscow", "count": "2"},
			wantStatus: http.StatusOK,
			wantCount:  2,
		},
		{
			label:      "очень много кафе для Москвы",
			reqParams:  map[string]string{"city": "moscow", "count": "100"},
			wantStatus: http.StatusOK,
			wantCount:  len(cafeList["moscow"]),
		},
	}

	for _, v := range tests {
		t.Run(v.label, func(t *testing.T) {

			reqString := buildRequest(v.reqParams)
			resp := getResponse(reqString)
			require.Equal(t, v.wantStatus, resp.status, ErrMsg, v.reqParams, resp.answer)

			require.Equal(t, v.wantCount, resp.count, ErrMsg, v.reqParams, resp.answer)

		})
	}
}
