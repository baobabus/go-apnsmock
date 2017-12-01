// Copyright 2017 Aleksey Blinov. All rights reserved.

package apns2

import (
	"strings"
	"testing"

	"github.com/baobabus/go-apnsmock/apns2mock"
)

func TestRoundtrip(t *testing.T) {
	s, err := apns2mock.NewServer(apns2mock.NoDelayCommsCfg, apns2mock.AllOkayHandler, apns2mock.AutoCert, apns2mock.AutoKey)
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()

	// client is preconfigured for communication with the server
	client := s.Client()
	url := s.URL + apns2mock.RequestPath
	cont := "application/json; charset=utf-8"

	// Expecting to get 200 back
	resp, err := client.Post(url, cont, strings.NewReader("{}"))
	if err != nil {
		t.Fatal(err)
	}
	if resp == nil {
		t.Fatal("Should have gotten a response")
	}
	if resp.StatusCode != 200 {
		t.Fatal("Should have gotten status 200")
	}

	s.BecomeUnavailable(503, "Shutdown")

	// Now expecting 503
	resp, err = client.Post(url, cont, strings.NewReader("{}"))
	if err != nil {
		t.Fatal(err)
	}
	if resp == nil {
		t.Fatal("Should have gotten a response")
	}
	if resp.StatusCode != 503 {
		t.Fatal("Should have gotten status 503")
	}
}
