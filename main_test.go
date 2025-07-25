package main

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"net/url"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func getResponse(reqString string, verbose bool) responseData {

	handler := http.HandlerFunc(mainHandle)

	resp := httptest.NewRecorder()
	req := httptest.NewRequest("GET", reqString, nil)

	handler.ServeHTTP(resp, req)

	if verbose {
		fmt.Printf("answer: %q\n", strings.TrimSpace(resp.Body.String()))
	}

	if resp.Code == http.StatusOK {
		if resp.Body.String() == "" {
			return responseData{resp.Code, 0, nil}
		} else {
			slice := strings.Split(resp.Body.String(), ",")

			for i := range slice{
				slice[i] = strings.TrimSpace(strings.ToLower(slice[i]))
			}

			return responseData{resp.Code, len(slice), slice}
		}
	}
	return responseData{resp.Code, 0, nil}
}

func buildRequest(r requestData, verbose bool) string {

	values := url.Values{}

	if r.city != nil {
		values.Set("city", *r.city)
	}

	if r.count != nil {
		values.Set ("count", *r.count)
	}

	if r.search != nil {
		values.Set("search", *r.search)
	}


	if verbose {
		fmt.Printf("query: %v\n",values)
	}

	query := values.Encode()

	if query =="" {
		return  "/cafe"
	}
	
	return "/cafe?" + query
	
}

func ptr (s string) *string {return &s}

type requestData struct {
	city  *string 
	count  *string
	search *string
}

type responseData struct {
	status int
	count  int
	slice  []string
}

type testItem struct {
	label      string
	req        requestData
	wantStatus int
	wantCount  int
}

func runTestCase(t *testing.T, tI testItem, verbose bool) {

	reqString := buildRequest(tI.req, verbose)
	resp := getResponse(reqString, verbose)
	require.Equal(t, tI.wantStatus, resp.status)
	assert.Equal(t, tI.wantCount, resp.count)

	if tI.req.search != nil {
		for _,v := range resp.slice {
			assert.Contains(t, strings.TrimSpace(v), strings.ToLower(*tI.req.search))
		}
	}
}



func TestCafeNegative(t *testing.T) {

	tests := []testItem{
		{
			"нe указано никаких параметров",
			requestData{nil, nil, nil},
			http.StatusBadRequest,
			0,
		},
		{
			"запрошен город Омск, по которому нет данных",
			requestData{ptr("Omsk"), nil, nil},
			http.StatusBadRequest,
			0,
		},
		{
			"указано некорректное значение параметра count для Тулы", 
			requestData{ptr("tula"), ptr("na"), nil},
			http.StatusBadRequest,
			0,
		},
	
	}

	for _, v := range tests {
		t.Run(v.label, func(t *testing.T) {
			runTestCase(t, v, false)
		})
	}

}

func TestCafeWhenOk(t *testing.T) {

	tests := []testItem{
		{
			"2 кафе для Москвы",
			requestData{ptr("moscow"), ptr("2"), nil},
			http.StatusOK,
			2,
		},
		{
			"указан только один параметр: город Тула",
			requestData{ptr("tula"), nil, nil},
			http.StatusOK,
			3,
		},
		{
			"ищем кафе с 'ложка' в Москве", 
			requestData{ptr("moscow"), nil, ptr("ложка")},
			http.StatusOK,
			1,
		},
	
	}

	for _, v := range tests {
		t.Run(v.label, func(t *testing.T) {
			runTestCase(t, v, false)
		})
	}

}



func TestCafeSearch(t *testing.T) {

	tests := []testItem{
		{
			"слово которого нет ни в одном кафе среди всех городов",
			requestData{ptr("moscow"), ptr("100"), ptr("фасоль")},
			http.StatusOK,
			0,
		},
		{
			"слово которое встречается 2 раза для Москвы",
			requestData{ptr("moscow"), ptr("100"), ptr("кофе")},
			http.StatusOK,
			2,
		},
		{
			"слово которое встречается 1 раз для Москвы", 
			requestData{ptr("moscow"), ptr("100"), ptr("вилка")},
			http.StatusOK,
			1,
		},
		{
			"слово которое встречается 1 раз для Тулы",
			requestData{ptr("tula"), ptr("100"), ptr("мир")},
			http.StatusOK,
			1,
		},
		{
			"пустой параметр search для Тула", 
			requestData{ptr("tula"), ptr("100"), ptr("")},
			http.StatusOK,
			3,
		},
		{
			"только цифры в парметре search для Москвы",
			requestData{ptr("moscow"), ptr("100"), ptr("1812")},
			http.StatusOK,
			0,
		},
		{
			"параметр search отсутствует в запросе",
			requestData{ptr("moscow"), ptr("100"), nil},
			http.StatusOK,
			5,
		},

		{
			"задан пустой параметр search",
			requestData{ptr("moscow"), ptr("100"), ptr("")},
			http.StatusOK,
			5,
		},
	}

	for _, v := range tests {
		t.Run(v.label, func(t *testing.T) {
			runTestCase(t, v, true)
		})
	}

}

func TestCafeCount(t *testing.T) {

	tests := []testItem{
		{
			"0 кафе для Москвы",
			requestData{ptr("moscow"), ptr("0"), ptr("")},
			http.StatusOK,
			0,
		},
		{
			"1 кафе для Москвы",
			requestData{ptr("moscow"), ptr("1"), ptr("")},
			http.StatusOK,
			1,
		},
		{
			"2 кафе для Тулы", 
			requestData{ptr("tula"), ptr("2"), ptr("")},
			http.StatusOK,
			2,
		},
		{
			"очень много кафе для Москвы",
			requestData{ptr("moscow"), ptr("100"), ptr("")},
			http.StatusOK,
			5,
		},
		{
			"4 кафе для Тулы", 
			requestData{ptr("tula"), ptr("4"), ptr("")},
			http.StatusOK,
			3,
		},
		{
			"задан пустой параметр count",
			requestData{ptr("moscow"), ptr(""), ptr("")},
			http.StatusOK,
			5,
		},
		{
			"параметр count отсутствует в запросе",
			requestData{ptr("moscow"), nil, ptr("")},
			http.StatusOK,
			5,
		},
	}

	for _, v := range tests {
		t.Run(v.label, func(t *testing.T) {
			runTestCase(t, v, false)
		})
	}

}
