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

type ErrorStruct400 struct {
	Message string `json:"error"`
}
