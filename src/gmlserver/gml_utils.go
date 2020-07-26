package gmlserver

import (
        "bytes"
        "encoding/json"
        "crypto/sha256"
        "io/ioutil"
        "net/http"
        "fmt"
        "strings"
)

func generateCodeVerifierAndCaculateCodeChallenge() (string, string, error) {
        // generate code_verifier
        codeVerifier, err := GenerateRandomString(32)
        if err != nil {
                return "", "", err
        }

        // BASE64URL-ENCODE(SHA256(ASCII("code_verifier" ))) == code_challenge
        h := sha256.New()
        _, err = h.Write([]byte(codeVerifier))
        if err != nil {
                return "", "", err
        }

        codeChallenge := Base64URLEncode(h.Sum(nil))

        return codeVerifier, codeChallenge, nil
}

func SendRequestAndCheckResponse(requestMethod string, request interface{}, expectedStatus int, expectedStruct interface{}) error {
        var payload []byte
        var err error
        switch request.(type) {
        case string:
                payload = []byte(request.(string))
        case []byte:
                payload = request.([]byte)
        default:
                payload, err = json.Marshal(request)
                if err != nil {
                        return fmt.Errorf("unable to marshal request to %s, reason: %s", requestMethod, err)
                }
        }

        resultBody, err := SendRequestToSimServer(requestMethod, payload, expectedStatus)
        if err != nil {
                return err
        }

        if expectedStruct != nil {
                err = json.Unmarshal(resultBody, &expectedStruct)
                if err != nil {
                        return fmt.Errorf("handler returned unexpected body: could not unmarshal into the structure we were expecting :: %v", err)
                }
        }
        return nil
}


func SendRequestToSimServer(requestMethod string, request []byte, expectedStatus int) ([]byte, error) {
        // make requestMethod lowercase as per our simulator server convention
        requestMethod = strings.ToLower(requestMethod)

        var result *http.Response
        var err error
        //log.Printf("--> send POST request to %s/%s, request body: %s\n", Config.SimServerURL, requestMethod, string(request))
        result, err = http.DefaultClient.Do(BuildRequest(http.MethodPost, Config.SimServerURL+"/"+requestMethod, request))
        if err != nil {
                return nil, fmt.Errorf("error sending request to Simulator Server :: %v", err)
        }

        defer result.Body.Close()
        response, err := ioutil.ReadAll(result.Body)
        if err != nil {
                return nil, fmt.Errorf("could not read response body :: %v", err)
        }
        //log.Printf("<-- received response from simulator server: %s\n\n", string(response))

        if result.StatusCode != expectedStatus {
                return nil, fmt.Errorf("handler returned wrong status code: got %v want %v, %s", result.StatusCode, expectedStatus, string(response))
        }
        return response, nil
}

func BuildRequest(method string, url string, payload []byte) *http.Request {
        req, _ := http.NewRequest(method, url, bytes.NewReader(payload))
        req.Header.Add("content-type", "application/json; charset=UTF-8")
        req.Header.Add("cache-control", "no-cache")
        // set connection: close header to disable keepalive
        // without this header the request may fail with error message "http: server closed idle connection" when server decide to reset TCP connection
        req.Header.Add("Connection", "close")
        return req
}
