package gmlserver

import (
	"encoding/json"
	"fmt"
	"net/http"
)

// retrieveCurrentTerms returns the ServerState with current terms and conditions updated
func RetrieveCurrentTerms(accessToken string) (string, error) {
	payload := new(RetrieveCurrentTermsReq)
	payload.Body.RetrieveCurrentTermsBody = &RetrieveCurrentTermsReqBody{
		AccessToken: accessToken,
		Endpoint:    Config.MyBankBaseURL,
		Locale:      "en-CA",
	}

	payloadBytes, err := json.Marshal(payload.Body)
	if err != nil {
		return "", fmt.Errorf("retrieveCurrentTerms: error when unmarshal request body:: %v", err)
	}

	// this checks expected status
	result := new(RetrieveCurrentTermsResp)
	err = SendRequestAndCheckResponse(RequestMethodRetrieveCurrentTerms, payloadBytes, http.StatusAccepted, &result.Body)
	if err != nil {
		return "", fmt.Errorf("retrieveCurrentTerms: error when SendRequestToSimServer:: %v", err)
	}

	return result.Body.ServerState, nil
}

// CreateLockboxWithOptionalRecoveryData ...
func CreateLockboxWithOptionalRecoveryData(accessToken string, withRecoveryData bool) (serverState string, err error) {
	if accessToken == "" {
		return "", fmt.Errorf("CreateLockboxWithOptionalRecoveryData -> cannot create Lockbox, must call getAuthToken first")
	}

	state, err := RetrieveCurrentTerms(accessToken)
	if err != nil {
		return "", fmt.Errorf("CreateLockboxWithOptionalRecoveryData->RetrieveCurrentTerms: %v", err)
	}
	payload := &CreateLockboxReqBody{
		AccessToken:             accessToken,
		Endpoint:                Config.MyBankBaseURL,
		ServerState:             state,
		DoNotCreateRecoveryData: !withRecoveryData,
	}

	var postbody = new(CreateLockboxReq)
	postbody.Body.CreateLockBoxbody = payload

	expected := new(CreateLockboxResp)
	if err = SendRequestAndCheckResponse(EndpointCreateLockbox, postbody.Body, http.StatusAccepted, &expected.Body); err != nil {
		return "", fmt.Errorf("CreateLockboxWithOptionalRecoveryData: error calling simulator server :: %#v", err)
	}
	if (CreateLockboxResp{}) == *expected {
		return "", fmt.Errorf("CreateLockboxWithOptionalRecoveryData: error calling simulator server :: the response is zero value")
	}

	return expected.Body.ServerState, nil
}
