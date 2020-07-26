package gmlserver

import (
        "net/http"
_        "encoding/json"
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
