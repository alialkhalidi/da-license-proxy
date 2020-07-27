package gmlserver

import (
	"strings"
	"time"
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
	err := SendRequestAndCheckResponse(strings.ToLower(RequestMethodRecoverLockbox), req.Body, expectedStatus, &expected.Body)
	if err != nil {
		return "", nil, err
	}
	return expected.Body.ServerState, expected.Body.RecoverLockboxBody, nil

}
