package tester

import (
	"context"
	"io/ioutil"
	"net"
	"net/http"
	"net/http/cookiejar"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

// DoRequestOnHandler does a request on a router (or handler) and check the response
func DoRequestOnHandler(t *testing.T, router http.Handler, method string, route string, authHeader string, payload string, expectedStatus int, expectedBody string) string {
	req, err := http.NewRequest(method, route, strings.NewReader(payload))
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Authorization", authHeader)
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)
	if status := rr.Code; status != expectedStatus {
		t.Errorf("Tested %v %v %v ; handler returned wrong status code: got %v want %v", method, route, payload, status, expectedStatus)
	}
	if !strings.HasPrefix(rr.Body.String(), expectedBody) {
		t.Errorf("Tested %v %v %v ; handler returned unexpected body: got %v want %v", method, route, payload, rr.Body.String(), expectedBody)
	}
	return string(rr.Body.String())
}

// DoRequestOnServer does a request on listening server
func DoRequestOnServer(t *testing.T, hostname string, port string, jar *cookiejar.Jar, method string, url string, authHeader string, payload string, expectedStatus int, expectedBody string) string {
	dialer := &net.Dialer{
		Timeout:   30 * time.Second,
		KeepAlive: 30 * time.Second,
		DualStack: true,
	}
	// or create your own transport, there's an example on godoc.
	http.DefaultTransport.(*http.Transport).DialContext = func(ctx context.Context, network, addr string) (net.Conn, error) {
		addrAndPort := strings.Split(addr, ":")
		if strings.HasSuffix(addrAndPort[0], "vestibule.io") {
			addr = "127.0.0.1:" + addrAndPort[1]
		}
		return dialer.DialContext(ctx, network, addr)
	}
	if strings.HasPrefix(url, "/") {
		url = "http://" + hostname + ":" + port + url
	} else {
		url = "http://" + url + ":" + port
	}
	req, err := http.NewRequest(method, url, strings.NewReader(payload))
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Authorization", authHeader)
	client := &http.Client{Jar: jar}
	res, err := client.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		t.Fatal(err)
	}
	bodyString := string(body)
	if status := res.StatusCode; status != expectedStatus {
		t.Errorf("Tested %v %v %v ; handler returned wrong status code: got %v want %v", method, url, payload, status, expectedStatus)
	}
	if !strings.HasPrefix(bodyString, expectedBody) {
		t.Errorf("Tested %v %v %v ; handler returned unexpected body: got %v want %v", method, url, payload, bodyString, expectedBody)
	}
	return bodyString
}

// CreateServerTester wraps DoRequestOnServer to factorize t, port and jar
func CreateServerTester(t *testing.T, hostname string, port string, jar *cookiejar.Jar) func(method string, url string, authHeader string, payload string, expectedStatus int, expectedBody string) {
	return func(method string, url string, authHeader string, payload string, expectedStatus int, expectedBody string) {
		DoRequestOnServer(t, port, hostname, jar, method, url, authHeader, payload, expectedStatus, expectedBody)
	}
}
