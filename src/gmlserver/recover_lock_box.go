package gmlserver

import (
	"strings"
	"time"
)

const (
	retry_attempts = 3
	retry_backoff  = 10
)

func RecoverLockboxWithClientID(accessToken string, expectedStatus int, clientID string) (string, *RecoverLockboxRespBody, error) {
	payload := &RecoverLockboxReqBody{
		AccessToken: accessToken,
		Endpoint:    Config.MyBankBaseURL,
		Locale:      "en",
		ClientID:    clientID,
	}

	var req = new(RecoverLockboxReq)
	req.Body.RecoverLockBoxBody = payload
	// give a few seconds for pre-conditions to propagte
	time.Sleep(10 * time.Second)
	var expected = new(RecoverLockboxResp)
	retries := 0
retry:
	err := SendRequestAndCheckResponse(strings.ToLower(RequestMethodRecoverLockbox), req.Body, expectedStatus, &expected.Body)
	if err != nil {
		if strings.Contains(err.Error(), "504") {
			if retries < 3 {
				retries++
				myLogger.Printf("RecoverLockboxWithClientID: recieved gateway 504 timeout, retry attempt %d/%d with %ds backoff", retries, retry_attempts, retry_backoff)
				time.Sleep(retry_backoff * time.Second)
				goto retry
			}
		}
		return "", nil, err
	}
	return expected.Body.ServerState, expected.Body.RecoverLockboxBody, nil

}
