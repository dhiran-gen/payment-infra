package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"testing"
	"time"
)

type Payment struct {
	ID     string `json:"id"`
	Amount int    `json:"amount"`
}

func waitForServer(url string) bool {
	for i := 0; i < 30; i++ {
		resp, err := http.Get(url)
		if err == nil && resp.StatusCode == 200 {
			return true
		}
		time.Sleep(1 * time.Second)
	}
	return false
}

func TestPaymentAPI_E2E(t *testing.T) {
	baseURL := "http://localhost:8080"
	if !waitForServer(baseURL + "/healthz") {
		t.Fatal("API server not up after 30s")
	}

	// Use unique ID and amount for each test run
	ts := time.Now().UnixNano()
	uniqueID := fmt.Sprintf("test%d", ts)
	uniqueAmount := int(ts % 10000)

	// 1. Create payment
	p := Payment{ID: uniqueID, Amount: uniqueAmount}
	body, _ := json.Marshal(p)
	resp, err := http.Post(baseURL+"/payments", "application/json", bytes.NewReader(body))
	if err != nil || resp.StatusCode != 200 {
		msg, _ := ioutil.ReadAll(resp.Body)
		resp.Body.Close()
		t.Fatalf("Create payment failed: %v, status: %d, body: %s", err, resp.StatusCode, string(msg))
	}

	// 2. Get payment (should hit cache after first call)
	resp, err = http.Get(baseURL + "/payments/" + p.ID)
	if err != nil || resp.StatusCode != 200 {
		msg, _ := ioutil.ReadAll(resp.Body)
		resp.Body.Close()
		t.Fatalf("Get payment failed: %v, status: %d, body: %s", err, resp.StatusCode, string(msg))
	}
	var getResp map[string]interface{}
	_ = json.NewDecoder(resp.Body).Decode(&getResp)
	resp.Body.Close()
	if getResp["id"] != p.ID || int(getResp["amount"].(float64)) != p.Amount {
		t.Fatalf("Get payment wrong data: %+v", getResp)
	}

	// 3. List payments
	resp, err = http.Get(baseURL + "/payments")
	if err != nil || resp.StatusCode != 200 {
		msg, _ := ioutil.ReadAll(resp.Body)
		resp.Body.Close()
		t.Fatalf("List payments failed: %v, status: %d, body: %s", err, resp.StatusCode, string(msg))
	}
	var listResp []Payment
	_ = json.NewDecoder(resp.Body).Decode(&listResp)
	resp.Body.Close()
	found := false
	for _, pay := range listResp {
		if pay.ID == p.ID && pay.Amount == p.Amount {
			found = true
		}
	}
	if !found {
		t.Fatalf("Created payment not found in list")
	}

	// 4. Update payment
	updateAmount := uniqueAmount + 100
	update := map[string]int{"amount": updateAmount}
	body, _ = json.Marshal(update)
	req, _ := http.NewRequest(http.MethodPut, baseURL+"/payments/"+p.ID, bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp, err = http.DefaultClient.Do(req)
	if err != nil || resp.StatusCode != 200 {
		msg, _ := ioutil.ReadAll(resp.Body)
		resp.Body.Close()
		t.Fatalf("Update payment failed: %v, status: %d, body: %s", err, resp.StatusCode, string(msg))
	}
	resp.Body.Close()

	// 5. Get updated payment
	resp, err = http.Get(baseURL + "/payments/" + p.ID)
	if err != nil || resp.StatusCode != 200 {
		msg, _ := ioutil.ReadAll(resp.Body)
		resp.Body.Close()
		t.Fatalf("Get updated payment failed: %v, status: %d, body: %s", err, resp.StatusCode, string(msg))
	}
	_ = json.NewDecoder(resp.Body).Decode(&getResp)
	resp.Body.Close()
	if int(getResp["amount"].(float64)) != updateAmount {
		t.Fatalf("Update not reflected: %+v", getResp)
	}

	// 6. Delete payment
	req, _ = http.NewRequest(http.MethodDelete, baseURL+"/payments/"+p.ID, nil)
	resp, err = http.DefaultClient.Do(req)
	if err != nil || resp.StatusCode != 200 {
		msg, _ := ioutil.ReadAll(resp.Body)
		resp.Body.Close()
		t.Fatalf("Delete payment failed: %v, status: %d, body: %s", err, resp.StatusCode, string(msg))
	}
	resp.Body.Close()

	// 7. Get deleted payment (should 404)
	resp, err = http.Get(baseURL + "/payments/" + p.ID)
	if err != nil {
		t.Fatalf("Get deleted payment failed: %v", err)
	}
	if resp.StatusCode != 404 {
		b, _ := ioutil.ReadAll(resp.Body)
		resp.Body.Close()
		t.Fatalf("Expected 404, got %d, body: %s", resp.StatusCode, string(b))
	}
	resp.Body.Close()
}
