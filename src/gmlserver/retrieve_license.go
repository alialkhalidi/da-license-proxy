package gmlserver

import (
	"fmt"
	"net/http"
)

func RetrieveLicenseRequest(accessToken, state, licenseRequestID, requestEncKey string, expectedStatus int) (string, *RetrieveLicenseRequestResp, error) {
	if accessToken == "" || state == "" {
		return "", nil, fmt.Errorf("retrieveLicenseRequest -> cannot retrieve license, must call createLockbox first")
	}

	_, err := decodeSimState(state)
	if err != nil {
		return "", nil, err
	}

	payload := RetrieveLicenseRequestReqBody{
		AccessToken:      accessToken,
		ServerState:      state,
		Endpoint:         Config.MyBankBaseURL,
		LicenseRequestID: licenseRequestID,
		RequestEncKey:    requestEncKey,
	}
	request := new(RetrieveLicenseRequestReq)
	request.Body.RetrieveLicenseRequestBody = &payload
	expected := new(RetrieveLicenseRequestResp)
	err = SendRequestAndCheckResponse("retrievelicenserequest", request.Body, http.StatusAccepted, &expected.Body)
	if err != nil {
		return "", nil, err
	}
	_, err = decodeSimState(expected.Body.ServerState)
	if err != nil {
		return "", nil, err
	}

	return expected.Body.ServerState, expected, nil

}
