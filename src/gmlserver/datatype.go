package gmlserver

import (
	_ "encoding/json"
	"net/http"
	"net/url"
	"time"
)

const (
	correctState                = "state"
	createLockboxScope          = "openid lockbox_creation"
	VerifiedMeScope             = "openid lockbox_creation verified_me"
	authlevelAT                 = "https://verified.me/loa/can/auth/elevated" //testGetAccessToken
	authlevelCLB                = "https://verified.me/loa/can/auth/elevated" //testCreateLockbox
	authlevelGRF                = "https://verified.me/loa/can/auth/standard" //GetReferenceData
	requestObjectRequestMethod  = "requestobject"
	accessTokenRequestMethod    = "accesstoken"
	RequestMethodRecoverLockbox = "recoverLockbox"
	EndpointCreateDigitalAsset  = "createdigitalasset"
	defaultTimeout              = time.Minute * 10
)

var Config Configuration

type Configuration struct {
	InteractionTypesDAP   map[string]string `envconfig:"interaction_types_dap" required:"true"`
	MemberIDMap           map[string]string `envconfig:"member_id_map" required:"true"`
	MyBankBaseURL         string            `envconfig:"mybank_base_url" required:"true"`
	MyBankOriginalBaseURL string            `envconfig:"mybank_original_base_url" required:"true"`
	DapAdapterBaseURL     string            `envconfig:"dap_adapter_base_url" required:"true"`
	CorrectProviderURL    string            `envconfig:"oidc_issuer_url" required:"true"`
	CorrectAudience       string            `envconfig:"oidc_issuer_url" required:"true"`
	UILocales             string            `envconfig:"ui_locales" required:"true"`
	ServerAddressDAC      string            `envconfig:"server_address_dac" required:"true"`
	SimServerURL          string            `envconfig:"sim_server_address" required:"true"`
	Protocol              string            `envconfig:"protocol" default:"https://"`
	ParallelTests         bool              `envconfig:"parallel_tests"`
	MTDACAdapterURL       string            `envconfig:"mtdac_adapter_url"`
	R12DACAdapterURL      string            `envconfig:"r12dac_adapter_url"`
	MTDACList             []string          `envconfig:"mtdac_list"`
}

type AccessTokenReq struct {
	//in: body
	Body struct {
		// Access Token request body to retreive the accesstoken from the provider.
		AccessTokenBody *AccessTokenReqBody `json:"accessTokenBody"`
	}
}

type AccessTokenReqBody struct {
	// Provider URL is the OIDC auth endpoint
	//required: true
	Provider string `json:"provider_url" validate:"required"`
	// Audience is the intended recipient of the request object.
	//required: true
	Audience string `json:"aud"`
	// AuthCode retreived from provider after authentication and consent.
	//required: true
	AuthCode string `json:"authCode" validate:"required"`
	// Endpoint to redirect to after getting Acesss Token
	//required: true
	RedirectURL string `json:"redirectUrl" validate:"required"`
	// PKCE Code Verifier
	//required: false
	CodeVerifier string `json:"code_verifier"`
	// ClientID if empty this will default to the first client configured in the app simulator.
	ClientID string `json:"clientId,omitempty"`
}

type AccessTokenResp struct {
	//in: body
	Body struct {
		// AccessToken that was retreived from the provider
		AccessToken string `json:"accesstoken"`
		// IDtoken that was retreived from the provider
		IDToken string `json:"idtoken"`
	}
}

type RequestObjectReq struct {
	//in: body
	Body struct {
		// RequestObject request body.
		RequestObjBody *RequestObjectReqBody `json:"requestObjBody"`
	}
}

// struct used to read extra parameters in request object.
type requestObjectReqExtra struct {
	// RequestObject request body.
	RequestObjBody map[string]interface{} `json:"requestObjBody"`
}

type RequestObjectReqBody struct {
	// Provider is the openid auth source.
	//required: true
	Provider string `json:"provider_url"`
	// Audience is the intended recipient of the request object.
	//required: true
	Audience string `json:"aud"`
	// State value to be used for openid flow.
	//required: true
	State string `json:"state,omitempty"`
	// Scopes needed for authcode must be space seperated.
	//required: true
	Scopes string `json:"scopes"`
	// The requested Authentication Context Class Reference values.
	AcrValues string `json:"acr_values,omitempty"`
	// The end user's preferred languages.
	//required: true
	UILocales string `json:"ui_locales"`
	// Prompt optional OIDC prompt parameter (eg. prompt=login to force re-login)
	Prompt string `json:"prompt,omitempty"`
	// RedirectURL optional redirect_url, if set this URL will be used as the redirect url instead of the server pre-configured value
	RedirectURL string `json:"redirecturl"`
	// Code challenge.
	//required: false
	CodeChallenge string `json:"code_challenge,omitempty"`
	// Code challenge method.
	//required: false
	CodeChallengeMethod string `json:"code_challenge_method,omitempty"`
	// Nonce.
	//required: false
	Nonce string `json:"nonce,omitempty"`
	// ClientID if empty this will default to the first client configured in the app simulator.
	ClientID string `json:"clientId,omitempty"`
}

type RequestObjectResp struct {
	//in: body
	Body struct {
		// the loginurl generated by the simulator server
		LoginURL string `json:"loginurl"`
	}
}

type MyAMAuthenticator struct {
	userID       string
	password     string
	oidcAuthURL  *url.URL
	client       *http.Client
	authCode     string
	lastRedirect string
}

type RecoverLockboxReq struct {
	//in: body
	Body struct {
		// RecoverLockbox request body.
		RecoverLockBoxBody *RecoverLockboxReqBody `json:"recoverLockboxBody"`
	}
}

type RecoverLockboxReqBody struct {
	// AccessToken retreived from provider for specific scopes related to RecoverLockbox.
	//required: true
	AccessToken string `json:"accessToken" validate:"required"`
	// Endpoint to contact to initiate the RecoverLockbox flow
	//required: true
	Endpoint string `json:"endpoint" validate:"required"`
	// Locale to retrieve the terms and conditions for.
	//required: true
	Locale string `json:"locale" validate:"required"`
	// ClientID
	ClientID string `json:"clientId"`
	// NumberOfCodes optional request to DLBP to return a number of org codes for use with subsequent calls to DAP
	// required: false
	NumberOfCodes int `json:"numberOfCodes,omitempty"`
}

type RecoverLockboxResp struct {
	//in: body
	Body struct {
		// RecoverLockBoxBody full response from the endpoint
		RecoverLockboxBody *RecoverLockboxRespBody `json:"recoverLockBoxBody"`
		// base64url encoded server state for representing the device internal state
		ServerState string `json:"serverState"`
	}
}

type RecoverLockboxRespBody struct {
	CreateLockboxRespBody
	Assets        []CreateDigitalAssetRespBody `json:"assets,omitempty"`
	Pseudonyms    []RecoverLockboxPseudonym    `json:"pseudonyms,omitempty"`
	DacPseudonyms []recoverLockboxDacPseudonym `json:"dacPseudonyms,omitempty"`
	RecoveryData  *RecoveryInfo                `json:"recoveryData,omitempty"`
	TermsInfo     *TermsInfo                   `json:"termsInfo,omitempty"`
	getOrgCodesDeviceRespBody
}

type CreateLockboxReq struct {
	//in: body
	Body struct {
		// CreateLockbox request body.
		CreateLockBoxbody *CreateLockboxReqBody `json:"createLockboxBody"`
	}
}

type CreateLockboxReqBody struct {
	// AccessToken retrieved from provider for specific scopes related to CreateLockbox.
	//required: true
	AccessToken string `json:"accessToken" validate:"required"`
	// Endpoint to contact to initiate the CreateLockbox flow
	//required: true
	Endpoint string `json:"endpoint" validate:"required"`
	// Server State is the base64url encoded state representing the internal state of the device
	//required: true
	ServerState string `json:"serverState" validate:"required"`
	// doNotCreateRecoveryData optional flag to disable sending recovery info during createlockbox
	//required: false
	DoNotCreateRecoveryData bool `json:"doNotCreateRecoveryData"`
	// assetTypes optional flag to create specified digital assets together with createLockbox
	//required: false
	AssetTypes []string `json:"assetTypes,omitempty"`
	// NumberOfCodes optional request to DLBP to return a number of org codes for use with subsequent calls to DAP
	// required: false
	NumberOfCodes int `json:"numberOfCodes,omitempty"`
}

type CreateLockboxResp struct {
	//in: body
	Body struct {
		// CreateLockBoxBody full response from the endpoint
		CreateLockboxBody *CreateLockboxRespBody `json:"createLockBoxBody"`
		// base64url encoded server state for representing the device internal state
		ServerState string `json:"serverState"`
	}
}

// CreateLockboxRespBody .
type CreateLockboxRespBody struct {
	// User information
	User *UserCreateLockboxResponse `json:"user" validate:"required"`
	// Device information
	Device *DeviceCreateLockBoxResponse `json:"device" validate:"required"`
	// Pseudonym information
	Pseudonym *PseudonymCreateLockboxResponse `json:"pseudonym" validate:"required"`
	// PseudonymDevice information
	Pseudonymdevice *PseudonymDeviceCreateLockboxResponse `json:"pseudonymDevice" validate:"required"`
	// DeviceSecurityData Key id of base key used to derive enhanced login device security data for this device id.
	DeviceSecurityData string `json:"deviceSecurityData" validate:"required"`
	// CreatedAssets information of any assets that were created during lockbox creation
	CreatedAssets []CreateDigitalAssetRespBody `json:"createdAssets,omitempty"`
	//
	getOrgCodesDeviceRespBody
}

type PseudonymCreateLockboxResponse struct {
	// pseudonym id in base64
	ID string `json:"id" validate:"required"`
	// pseudonym derivation data in base64
	DerivationData string `json:"derivationData" validate:"required"`
	// the user id salt in base64
	UserIDSalt string `json:"userIdSalt" validate:"required"`
	// signing key derivation data in base64
	SigKeyDerivationData string `json:"sigKeyDerivationData" validate:"required"`
	// encryption key derivation data in base64
	EncKeyDerivationData string `json:"encKeyDerivationData" validate:"required"`
	// app encryption key derivation data in base64
	AppEncKeyDerivationData string `json:"appEncKeyDerivationData" validate:"required"`
	// id of dap being called
	MemberID string `json:"memberId" validate:"required"`
	// salt to prove member id
	MemberIDSalt string `json:"memberIdSalt" validate:"required"`
}

// PseudonymDeviceCreateLockboxResponse .
// swagger:model pseudonymDevice
type PseudonymDeviceCreateLockboxResponse struct {
	// Pseudonym device id in base64
	ID string `json:"id"`
	// Pseudonym id in base64
	PseudonymID string `json:"pseudonymId"`
	// Pseudonym id salt in base64
	PseudonymIDSalt string `json:"pseudonymIdSalt"`
	// Device id salt in base64
	DeviceIDSalt string `json:"deviceIdSalt"`
}

type UserCreateLockboxResponse struct {
	// User id in base64
	ID string `json:"id" validate:"required"`
	// Dlbp id in base64
	DlbpID string `json:"dlbpId" validate:"required"`
	// Dlbp id salt in base64
	DlbpIDSalt string `json:"dlbpIdSalt" validate:"required"`
	// Pseudonym derivation data in base64
	PseudonymBaseDerivationData string `json:"pseudonymBaseDerivationData" validate:"required"`
}

// DeviceCreateLockBoxResponse .
// swagger:model device
type DeviceCreateLockBoxResponse struct {
	// the device id in base64
	ID string `json:"id" validate:"required"`
	// the user id salt in base64
	UserIDSalt string `json:"userIdSalt" validate:"required"`
}

type getOrgCodesDeviceRespBody struct {
	Codes        []ChannelCodeWithExpiry `json:"codes"`
	CodeDuration int                     `json:"codeDuration"`
}

type ChannelCode struct {
	ID   string `json:"id" validate:"required"`
	HMAC string `json:"hmac" validate:"required"`
}

type ChannelCodeWithExpiry struct {
	ChannelCode
	Expiry int64 `json:"expiry" validate:"required"`
}

type CreateDigitalAssetRespBody struct {
	// asset id
	DigitalAssetID string `json:"digitalAssetId,omitempty"`
	// the asset type
	DigitalAssetType string `json:"digitalAssetType"`
	// pseudonym id
	PseudonymID string `json:"pseudonymId,omitempty"`
	// pseudonym id salt
	PseudonymIDSalt string `json:"pseudonymIdSalt,omitempty"`
	// asset base salt jwe. This is encrypted with the pseudonym encryption key
	AssetBaseSalt string `json:"assetBaseSalt,omitempty"`
	// asset base encryption key jwe. This is encrypted with the pseudonym encryption key
	AssetBaseEncryptionKey string `json:"assetBaseEncryptionKey,omitempty"`
	// type of storage
	StorageType string `json:"storageType,omitempty"`
	// digital asset provider id
	DapID string `json:"dapId,omitempty"`
	// expiry for the asset
	ExpiryEpochSeconds int64 `json:"expiryEpochSeconds,omitempty"`
	// salt that will use for creating license for this asset
	LicensedDigitalAssetIDSalt string `json:"licensedDigitalAssetIdSalt,omitempty"`
	// counter that keeps track of licenses issued for this asset
	LastSequenceNumber int `json:"lastSequenceNumber,omitempty"`
	// initial asset status
	Status string `json:"status,omitempty"`
	// estimated time until the asset is ACTIVE, if it is PENDING
	EstimatedActiveTime int64 `json:"estimatedActiveTime,omitempty"`
	// error
	Error *struct {
		Code        string `json:"code"`
		ErrorTest   string `json:"errorText,omitempty"`
		Recoverable bool   `json:"isRecoverable"`
	} `json:"error,omitempty"`
}

type RecoverLockboxPseudonym struct {
	createPseudonymResponsePseudonym
	PseudonymMemberJwe string `json:"pseudonymMemberJwe" validate:"required"`
}

type createPseudonymResponsePseudonym struct {
	// pseudonym id
	ID string `json:"id" validate:"required"`
	// derivation data used to derive pseudonym key
	DerivationData string `json:"derivationData" validate:"required"`
	// salt applied to generate pseudonym ID from user ID
	UserIDSalt string `json:"userIdSalt" validate:"required"`
	// derivation data used to generate signing key
	SigKeyDerivationData string `json:"sigKeyDerivationData" validate:"required"`
	// derivation data used to generate encryption key
	EncKeyDerivationData string `json:"encKeyDerivationData" validate:"required"`
	// derivation data used to generate app encryption key
	AppEncKeyDerivationData string `json:"appEncKeyDerivationData" validate:"required"`
}

type recoverLockboxDacPseudonym struct {
	ID          string `json:"id" validate:"required"`
	EncUserData string `json:"encUserData" validate:"required"`
	CreatedTime int64  `json:"createdTime" validate:"required"`
}

type RecoveryInfo struct {
	RecoveryData              recoveryDataField `json:"recoveryData" validate:"required"`
	EncDlbpRecoveryKeyPart    string            `json:"encDlbpRecoveryKeyPart" validate:"required"`
	EncStewardRecoveryKeyPart string            `json:"encStewardRecoveryKeyPart" validate:"required"`
}

type recoveryDataField struct {
	EncLockboxEncKey string `json:"encLockboxEncKey" validate:"required"`
}

type TermsInfo struct {
	Locale      string `json:"locale" validate:"required"`
	Version     string `json:"version" validate:"required"`
	ContentType string `json:"contentType" validate:"required"`
	Content     string `json:"content" validate:"required"`
}

type DLBstate struct {
	CreateLockboxResponse  CreateLockboxRespBody                 `json:"createLockBoxBody,omitempty"`
	RecoverLockboxResponse *RecoverLockboxRespBody               `json:"recoverLockboxResponse,omitempty"`
	RecoveryInfo           *DetailedRecoveryInfo                 `json:"detailedRecoveryInfo,omitempty"`
	MasterPrivateKey       string                                `json:"masterPrivateKey,omitempty"`
	PseudoDevicePrivateKey string                                `json:"pseudoDevicePrivateKey,omitempty"`
	DAList                 map[string]CreateDigitalAssetRespBody `json:"daList,omitempty"` /* asset-type-to-asset map */
	LastLicenseRequest     struct {
		LicenseRequestID  string             `json:"licenseRequestId"`
		DacLicenseRequest *DACLicenseRequest `json:"dacLicenseRequest"`
		RequestInfo       string             `json:"requestInfo"`
		PseudonymID       string             `json:"pseudonymId,omitempty"`
		PseudonymIDSalt   string             `json:"pseudonymIdSalt,omitempty"`
	} `json:"lastLicenseRequest,omitempty"`
	PseudonymMap                  map[string]*CreatePseudonymRespBody  `json:"pseudonymMap,omitempty"`
	DacPseudonymList              []dacPseudonym                       `json:"dacPseudonymList,omitempty"`            // an array of dac pseudonyms
	PseudonymDeviceKeyMap         map[string]string                    `json:"pseudoDeviceKeyMap,omitempty"`          // a map of pseudonym-deviceSigKey's
	PseudonymsClaimedPendingMap   map[string]*UserInteractionRequest   `json:"pseudonymsClaimedPendingMap,omitempty"` // a map of pseudonym claims pending on user interaction
	PseudonymsClaimedMap          map[string]*ClaimedPseudonym         `json:"pseudonymsClaimedMap,omitempty"`        // a map of successfully claimed pseudonyms
	PseudonymsInRecoveryMap       map[string]*RecoverLockboxPseudonym  `json:"pseudonymsInRecoveryMap,omitempty"`     // pseudonyms which can be recovered using recoverPseudonym at DAP
	ServiceResponses              map[string]executeServiceAdapterResp `json:"serviceResponses,omitempty"`
	CurrentTermsInfo              *termsAcceptance                     `json:"currentTermsInfo,omitempty"`
	AcceptedTerms                 *termsAcceptance                     `json:"acceptedTerms,omitempty"`
	GetTransactionHistoryResponse *GetTransactionHistoryRespBody       `json:"getTransactionHistoryResponse,omitempty"`
	ClientID                      string                               `json:"clientId,omitempty"`
	OrgCodes                      []ChannelCodeWithExpiry              `json:"orgCodes"`
}

type DetailedRecoveryInfo struct {
	RecoveryInfo
	RecoveryDataB64  string `json:"recoveryDataB64"`
	RecoveryDataHash string `json:"recoveryDataHash"`
	RecoveryDataSalt string `json:"recoveryDataSalt"`
	LockboxEncKey    string `json:"lockboxEncKey"`
	RecoveryKey      string `json:"recoveryKey"`
}

type DACLicenseRequest struct {
	// ID of the DAC(digital asset consumer).
	DacID string `json:"dacId"`
	// Salt used for blinding the ID of the DAC(digital asset consumer).
	DacIDSalt string `json:"dacIdSalt"`
	// DAC risk category ID
	DACRiskCategoryID string `json:"dacRiskCategoryId"`
	// DAC risk category ID salt
	DACRiskCategoryIDSalt string `json:"dacRiskCategoryIdSalt"`
	// Salt used for blinding the current license request.
	RequestSalt string `json:"requestSalt"`
	// Lists  asset types and fields that the DAC(digital asset consumer) wants.
	QueryExpression string `json:"queryExpression,omitempty"`
	// Salt used for blinding the QueryExpression.
	QueryExpressionSalt string `json:"queryExpressionSalt"`
	// A DAC(digital asset consumer) endpoint that device sends the license to.
	LicenseNotificationURL string `json:"licenseNotificationUrl"`
	// The License content will be encrypted by this public key provided by the DAC(digital asset consumer).
	LicenseEncKey string `json:"licenseEncKey"`
	// Text that will be displayed by the device app to the user as a part of the license.
	DisplayText interface{} `json:"displayText"`
	// Default language.
	DefaultLang string `json:"defaultLang"`
	// Encryption key for services.
	ServiceEncKey string `json:"serviceEncKey"`
	// An opaque value put into license by the DAC(digital asset consumer).
	State string `json:"state"`
	// Salt used for blinding the state.
	StateSalt string `json:"stateSalt"`
	Auth      *struct {
		ReturnID bool `json:"returnId"`
	} `json:"auth"`
	// DacIDSalts - a map of asset/serviceName to dacIdSalt
	DacIDSalts map[string]string `json:"dacIdSalts,omitempty"`
	// DacIDsHashesBase64 - base64Url encoded json string of blinded dac id map
	DacIDsHashesBase64 string `json:"dacIdsHashesBase64,omitempty"`
	// PseudonymID - optional pseudonymid to peg license request to a lockbox
	PseudonymID string `json:"pseudonymId,omitempty"`
	// PseudonymIDSalt - optional salt to hash pseudonym id with
	PseudonymIDSalt string `json:"pseudonymIdSalt,omitempty"`
}

type CreatePseudonymRespBody struct {
	Pseudonym       createPseudonymResponsePseudonym `json:"pseudonym" validate:"required"`
	PseudonymDevice struct {
		// the unique id of the pseudonym device
		ID string `json:"id" validate:"required"`
		// salt applied to generate the unique id of the pseudonym device
		PseudonymIDSalt string `json:"pseudonymIdSalt" validate:"required"`
	} `json:"pseudonymDevice" validate:"required"`
}

type dacPseudonym struct {
	ID            string `json:"id" validate:"required"`
	BlindedUserID string `json:"blindedUserId,omitempty"`
	BlindedDacID  string `json:"blindedDacId,omitempty"`
	DacID         string `json:"dacId,omitempty"`
	DacIDSalt     string `json:"dacIdSalt,omitempty"`
	CreatedTime   int64  `json:"createdTime" validate:"required"`
	EncUserData   string `json:"encUserData,omitempty"`
}

type UserInteractionRequest struct {
	UserInteractionURL   string `json:"userInteractionUrl,omitempty"`
	LicenseRequestEncKey string `json:"licenseRequestEncKey,omitempty"`
	AppHostState         string `json:"appHostState,omitempty"`
}

type ClaimedPseudonym struct {
	ID           string `json:"id" validate:"required"`
	MemberID     string `json:"memberId" validate:"required"`
	MemberIDSalt string `json:"memberIdSalt" validate:"required"`
}

type executeServiceAdapterResp struct {
	SuccessResponse *executeServiceAdapterSuccessResponse `json:"successResponse"`
	ErrorResponse   *executeServiceAdapterErrorResponse   `json:"errorResponse"`
}

type executeServiceAdapterSuccessResponse struct {
	ServiceResponseID      string                  `json:"serviceResponseId" validate:"required"`
	ServiceResponseEncKey  string                  `json:"serviceResponseEncKey" validate:"required"`
	ServiceResponseSalt    string                  `json:"serviceResponseSalt" validate:"required"`
	UserData               map[string]displayData  `json:"userData" validate:"min=1,dive,required"`
	UserInteractionRequest *UserInteractionRequest `json:"userInteractionRequest,omitempty"`
	PseudonymID            string                  `json:"pseudonymId" validate:"required"`
	PseudonymIDSalt        string                  `json:"pseudonymIdSalt" validate:"required"`
	LicenseRequestID       string                  `json:"licenseRequestId" validate:"required"`
	LicenseRequestIDSalt   string                  `json:"licenseRequestIdSalt" validate:"required"`
	FieldsSalt             string                  `json:"fieldsSalt"`
}
type executeServiceAdapterErrorResponse struct {
	Code        string          `json:"code"`
	Description textDescription `json:"description"`
}

type textDescription struct {
	Locale string `json:"locale"`
	Text   string `json:"text"`
}

type displayData struct {
	Type         string                 `json:"type" validate:"required,oneof=object field"`
	Fields       map[string]displayData `json:"fields,omitempty" validate:"dive"`
	Value        interface{}            `json:"value,omitempty"`
	DoNotDisplay bool                   `json:"doNotDisplay,omitempty"`
}

type termsAcceptance struct {
	Locale      string `json:"locale"`
	Version     string `json:"version"`
	ContentHash string `json:"contentHash"`
}

type GetTransactionHistoryRespBody struct {
	TransactionHistoryItems []TxHistoryItem `json:"transactionHistory,omitempty"`
	// MissingEvents - only for v1
	MissingEvents []string `json:"missingEvents,omitempty"`
	// LastEventNo - only for v1
	LastEventNo *uint32 `json:"lastEventNo,omitempty"`
}

type TxHistoryItem struct {
	ID string `json:"id" validate:"required"`
	// EventNo - only for header mode and V1
	EventNo *uint32 `json:"eventNo,omitempty"`
	// PseudonymID - only for header mode and V1
	PseudonymID string `json:"pseudonymId,omitempty"`
	// UserEventID - only for header mode and V1
	UserEventID string `json:"userEventId,omitempty"`
	// Date - only for header mode
	Date int64 `json:"date,omitempty"`
	// Data - only for data mode and V1
	Data string `json:"data,omitempty"`
	// Status - only for data mode
	Status string `json:"status,omitempty"`
}

type CreateDigitalAssetReq struct {
	//in: body
	Body struct {
		// createDA request body
		CreateDigitalAssetBody *CreateDigitalAssetReqBody `json:"CreateDigitalAssetBody"`
	}
}

type CreateDigitalAssetReqBody struct {
	// AccessToken retrieved from provider for specific scopes related to createDA.
	//required: true
	AccessToken string `json:"accessToken" validate:"required"`
	// Endpoint to contact to initiate the createDA flow.
	//required: true
	Endpoint string `json:"endpoint" validate:"required"`
	// ChannelCode The channel code for the DAP to use. If in a message to the DLBP then this is optional.
	// required: false
	ChannelCode *ChannelCode `json:"channelCode,omitempty"`
	// PseudonymID under which the assets will be created, can be omitted if this is for the owner pseudonymID of the lockbox
	PseudonymID string `json:"pseudonymId"`
	// Array of Asset type identifiers being created.
	//required: true
	AssetTypes []string `json:"assetTypes" validate:"min=1,dive,required"`
	// Asset status to inform demo daps to create asset(s) with given status
	//required: false
	AssetStatus string `json:"assetStatus" validate:"omitempty,oneof=ACTIVE PENDING REVOKED"`
	// UserInteractionInfo user interaction information to complete the user interaction before asset can be created
	UserInteractionInfo *UserInteractionInfo `json:"userInteractionInfo"`
	// License created from the encryption key returned by createDA call
	License *string `json:"license"`
	// AppHostState opaque state returned from apphost
	AppHostState *string `json:"appHostState"`
	// UILocales sets the preferred order of locales for the display page
	// required: false
	UILocales string `json:"ui_locales"`
	// Server State is the base64url encoded state representing the internal state of the device
	//required: true
	ServerState string `json:"serverState" validate:"required"`
}

type CreateDigitalAssetResp struct {
	//in: body
	Body struct {
		// createDA full response from endpoint.
		CreateDigitalAssetBody []CreateDigitalAssetRespBody `json:"createDigitalAssetBody"`
		// LicenseEncKey a license encryption key returned from endpoint, this indicates that the user must CreateDigitalAssetBody a license before proceeding to call createDA.
		LicenseEncKey *string `json:"licenseEncKey"`
		// AppHostState opaque state returned from apphost
		AppHostState *string `json:"appHostState"`
		// UserInteractionRequest User interaction info returned from endpoint. This indicates that the user must complete some user interaction in order to successfully CreateDigitalAssetBody the digital asset, be it visiting a URL or creating a license with a given license encryption key.
		UserInteractionRequest *UserInteractionRequest `json:"userInteractionRequest"`
		// base64url encoded server state for representing the state of the current device
		ServerState string `json:"serverState"`
	}
}

type UserInteractionInfo struct {
	URLReturnValue string `json:"urlReturnValue,omitempty"`
	License        string `json:"license,omitempty"`
	AppHostState   string `json:"appHostState,omitempty"`
}
