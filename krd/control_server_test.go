package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/agrinman/kr"
)

func TestControlServerPair(t *testing.T) {
	transport := &kr.ResponseTransport{T: t}
	ec := NewTestEnclaveClient(transport)
	cs := ControlServer{ec}
	pairRequest, err := http.NewRequest("PUT", "/pair", nil)
	if err != nil {
		t.Fatal(err)
	}
	recorder := httptest.NewRecorder()
	cs.handlePair(recorder, pairRequest)
	resp := recorder.Result()
	if resp.StatusCode != http.StatusOK {
		t.Fatal("non-200 status")
	}
	var pairingSecret kr.PairingSecret
	err = json.NewDecoder(resp.Body).Decode(&pairingSecret)
	if err != nil {
		t.Fatal(err)
	}

	getPairRequest, err := http.NewRequest("GET", "/pair", nil)
	if err != nil {
		t.Fatal(err)
	}
	recorder = httptest.NewRecorder()
	cs.handlePair(recorder, getPairRequest)
	resp = recorder.Result()
	if resp.StatusCode != http.StatusOK {
		t.Fatal("non-200 status")
	}
	var me kr.Profile
	err = json.NewDecoder(resp.Body).Decode(&me)
	if err != nil {
		t.Fatal(err)
	}
	testMe, _, _ := kr.TestMe(t)
	if !me.Equal(testMe) {
		t.Fatal("paired profile wrong")
	}
}

func TestControlServerUnpair(t *testing.T) {
	transport := &kr.ResponseTransport{T: t}
	ec := NewTestEnclaveClientShortTimeouts(transport)
	cs := ControlServer{ec}
	pairRequest, err := http.NewRequest("PUT", "/pair", nil)
	if err != nil {
		t.Fatal(err)
	}
	recorder := httptest.NewRecorder()
	cs.handlePair(recorder, pairRequest)
	resp := recorder.Result()
	if resp.StatusCode != http.StatusOK {
		t.Fatal("non-200 status")
	}
	var pairingSecret kr.PairingSecret
	err = json.NewDecoder(resp.Body).Decode(&pairingSecret)
	if err != nil {
		t.Fatal(err)
	}

	unpairRequest, err := http.NewRequest("DELETE", "/pair", nil)
	if err != nil {
		t.Fatal(err)
	}
	recorder = httptest.NewRecorder()
	cs.handlePair(recorder, unpairRequest)
	resp = recorder.Result()
	if resp.StatusCode != http.StatusOK {
		t.Fatal("non-200 status")
	}
	if ec.IsPaired() {
		t.Fatal("client should be unpaired")
	}

	getPairRequest, err := http.NewRequest("GET", "/pair", nil)
	if err != nil {
		t.Fatal(err)
	}
	recorder = httptest.NewRecorder()
	cs.handlePair(recorder, getPairRequest)
	resp = recorder.Result()
	if resp.StatusCode != http.StatusNotFound {
		t.Fatal("expected 404 not found")
	}
}

func TestControlServerMe(t *testing.T) {
	transport := &kr.ResponseTransport{T: t}
	ec := NewTestEnclaveClient(transport)
	cs := ControlServer{ec}
	request, err := kr.NewRequest()
	if err != nil {
		t.Fatal(err)
	}
	request.MeRequest = &kr.MeRequest{}

	meRequest, err := request.HTTPRequest()
	if err != nil {
		t.Fatal(err)
	}
	recorder := httptest.NewRecorder()
	cs.handleEnclave(recorder, meRequest)
	resp := recorder.Result()
	if resp.StatusCode != http.StatusNotFound {
		t.Fatal("expected 404, not paired")
	}

	pairClient(t, ec)
	defer ec.Stop()

	meRequest, err = request.HTTPRequest()
	if err != nil {
		t.Fatal(err)
	}
	recorder = httptest.NewRecorder()
	cs.handleEnclave(recorder, meRequest)
	resp = recorder.Result()
	if resp.StatusCode != http.StatusOK {
		t.Fatal("non-200 status")
	}

	var meResponse kr.Response
	err = json.NewDecoder(resp.Body).Decode(&meResponse)
	if err != nil {
		t.Fatal(err)
	}
	me, _, _ := kr.TestMe(t)
	if !meResponse.MeResponse.Me.Equal(me) {
		t.Fatal("profiles unequal")
	}
}

func TestControlServerSign(t *testing.T) {
	transport := &kr.ResponseTransport{T: t}
	ec := NewTestEnclaveClient(transport)
	cs := ControlServer{ec}
	request, err := kr.NewRequest()
	if err != nil {
		t.Fatal(err)
	}

	me, _, _ := kr.TestMe(t)
	digest, err := kr.RandNBytes(32)
	request.SignRequest = &kr.SignRequest{
		PublicKeyFingerprint: me.PublicKeyFingerprint(),
		Digest:               digest,
	}

	signRequest, err := request.HTTPRequest()
	if err != nil {
		t.Fatal(err)
	}
	recorder := httptest.NewRecorder()
	cs.handleEnclave(recorder, signRequest)
	resp := recorder.Result()
	if resp.StatusCode != http.StatusNotFound {
		t.Fatal("expected 404, not paired")
	}

	pairClient(t, ec)
	defer ec.Stop()

	signRequest, err = request.HTTPRequest()
	if err != nil {
		t.Fatal(err)
	}
	recorder = httptest.NewRecorder()
	cs.handleEnclave(recorder, signRequest)
	resp = recorder.Result()
	if resp.StatusCode != http.StatusOK {
		t.Fatal("non-200 status")
	}

	var signResponse kr.Response
	err = json.NewDecoder(resp.Body).Decode(&signResponse)
	if err != nil {
		t.Fatal(err)
	}
}

func TestControlServerPing(t *testing.T) {
	transport := &kr.ResponseTransport{T: t}
	ec := NewTestEnclaveClient(transport)
	cs := ControlServer{ec}
	pingRequest, err := http.NewRequest("GET", "/ping", nil)
	if err != nil {
		t.Fatal(err)
	}
	recorder := httptest.NewRecorder()
	cs.handlePing(recorder, pingRequest)
	resp := recorder.Result()
	if resp.StatusCode != http.StatusOK {
		t.Fatal("non-200 status")
	}
}
