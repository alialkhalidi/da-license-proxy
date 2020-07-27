package gmlserver

import (
	"fmt"
	"net/http"
)

func IssueLicense(accessToken string, state string, licenseRequestID string, daMap map[string]CreateDigitalAssetRespBody) (*IssueLicenseResp, error) {

	if accessToken == "" || state == "" {
		return nil, fmt.Errorf("IssueLicense -> cannot issue license, must call createLockbox first")
	}
	if licenseRequestID == "" {
		return nil, fmt.Errorf("IssueLicense -> cannot issue license, must call retrieve license first")
	}
	if daMap == nil {
		return nil, fmt.Errorf("IssueLicense -> cannot issue license, must call CreateDA first")
	}

	//check to see if assets have already been created
	_, err := decodeSimState(state)
	if err != nil {
		return nil, err
	}

	issueLicensePayload := &IssueLicenseReqBody{
		AccessToken:       accessToken,
		ServerState:       state,
		Endpoint:          Config.MyBankBaseURL,
		LicenseRequestID:  licenseRequestID,
		EncryptWholeAsset: true,
		DoNotNotifyDAC:    true,
	}

	issueLicenseReq := new(IssueLicenseReq)
	issueLicenseReq.Body.IssueLicenseBody = issueLicensePayload
	matchedAssets := make(map[string]AssetQueryEntry)
	matchedAssets["foundationalIdentityName"] = AssetQueryEntry{
		AssetSeqNo:     1,
		DigitalAssetID: daMap["vme://assets/foundationalIdentity"].DigitalAssetID,
		Name:           "asset1",
	}
	issueLicenseReq.Body.MatchedAssets = matchedAssets
	var expected = new(IssueLicenseResp)
	err = SendRequestAndCheckResponse(EndpointIssueLicense, issueLicenseReq.Body, http.StatusAccepted, &expected.Body)
	if err != nil {
		return nil, fmt.Errorf("%v", err)
	}
	return expected, nil

}
