package gmlserver

import (
	"encoding/json"
	"fmt"
	"net/http"
)

func decodeSimState(b64state string) (*DLBstate, error) {
	if b64state == "" {
		return &DLBstate{}, nil
	}
	decodedstate, _ := Base64URLDecode(b64state)
	statestruct := new(DLBstate)
	err := json.Unmarshal(decodedstate, statestruct)
	if err != nil {
		return nil, err
	}
	return statestruct, nil
}

func CreateDA(accessToken string, state string, assetTypes []string) (string, map[string]CreateDigitalAssetRespBody, error) {
	if accessToken == "" || state == "" {
		return "", nil, fmt.Errorf("createDA -> cannot create DA, must call createLockbox first")
	}
	//check to see if assets have already been created
	stateObj, err := decodeSimState(state)
	if err != nil {
		return "", nil, err
	}
	/*
		containsExactAssets := func(s *DLBstate, assetTypes []string) bool {
			//		if s.DAList == nil || len(s.DAList) == 0 || len(s.DAList) != len(assetTypes) {
			if s.DAList == nil || len(s.DAList) == 0 || len(s.DAList) >= len(assetTypes) {
				return false
			}
			contains := true
			for _, assetType := range assetTypes {
				if _, ok := s.DAList[assetType]; !ok {
					contains = false
				}
			}
			return contains
		}
		if containsExactAssets(stateObj, assetTypes) {
			return state, stateObj.DAList, nil
		}
	*/
	payload := &CreateDigitalAssetReqBody{
		AccessToken: accessToken,
		Endpoint:    Config.MyBankBaseURL,
		AssetTypes:  assetTypes,
		ServerState: state,
	}
	myLogger.Printf("Sending CreateDA, endpoint: %s, channelCode: %v\n", payload.Endpoint, payload.ChannelCode)

	var postbody = new(CreateDigitalAssetReq)
	postbody.Body.CreateDigitalAssetBody = payload
	expected := new(CreateDigitalAssetResp)
	if err = SendRequestAndCheckResponse(EndpointCreateDigitalAsset, postbody.Body, http.StatusAccepted, &expected.Body); err != nil {
		return "", nil, fmt.Errorf("sending of createDA request failed to %s", err)
	}

	if len(expected.Body.CreateDigitalAssetBody) != len(assetTypes) {
		return "", nil, fmt.Errorf("expected createDA to create %d assets", len(assetTypes))
	}

	stateObj, err = decodeSimState(expected.Body.ServerState)
	if err != nil {
		return "", nil, err
	}
	return expected.Body.ServerState, stateObj.DAList, nil
}
