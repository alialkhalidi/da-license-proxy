package gmlserver

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"

	"sk-git.securekey.com/gerrit/vme-core/utils/config"

	"github.com/spf13/viper"
)

var myLogger = log.New(log.Writer(), "gmlserver ", 0)

const (
	SERVER_ADDRESS = "http.url"
	SERVER_PORT    = "http.listen.address"
	SERVER_UI_PATH = "http.ui.path"
	SIMSERVER_URL  = "simserver.url"
	MYAM_URL       = "myam.url"
)

type GmlServer struct {
	ServerAddress string
	Port          string
	UIPath        string
	SimServerURL  string
	MyamURL       string
	listener      net.Listener
	writer        http.ResponseWriter
}

func NewGmlServer(cfgFile string) (*GmlServer, error) {
	instance := GmlServer{}
	err := instance.initConfig(cfgFile)

	return &instance, err

}

func getLicenseForDA(username, password, licenseRequestID, requestEncKey string) (string, error) {

	accessToken, err := GetAccessToken(VerifiedMeScope, username, password, "")
	if err != nil {
		myLogger.Printf("GetAccessToken: %v", err)
		return "", err
	}
	serverState, _, err := RecoverLockboxWithClientID(accessToken, http.StatusAccepted, "")
	if err != nil {
		myLogger.Printf("RecoverLockboxWithClientID: %v", err)
		return "", err
	}

	assets := []string{"vme://assets/foundationalIdentity"}
	serverState, daMap, err := CreateDA(accessToken, serverState, assets)
	if err != nil {
		myLogger.Printf("CreateDA: %v", err)
		return "", err
	}

	serverState, _, err = RetrieveLicenseRequest(accessToken, serverState, licenseRequestID, requestEncKey, http.StatusAccepted)
	if err != nil {
		myLogger.Printf("RetrieveLicenseRequest: %v", err)
		return "", err
	}

	issueLicenseResp, err := IssueLicense(accessToken, serverState, licenseRequestID, daMap)
	if err != nil {
		myLogger.Printf("IssueLicense: %v", err)
		return "", err
	}
	return issueLicenseResp.Body.License, nil

}

func (t *GmlServer) processGetMethod() {
	const page = `<html>
  <form id="gml" action="/ui" method="post">
  <textarea name="JSON" id="JSON" placeholder='{"username": "", "password": "", "requestId": "", "requestEncKey": ""}' spellcheck="false" rows="20" form="gml">
	</textarea>
  <input type="submit" value="Send Request<"/>
  </form>
  <html>
  `
	fmt.Fprintf(t.writer, "%s", page)
}

func (t *GmlServer) processPostMethod(r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		myLogger.Printf("processPostMethod: error parsing form")
		t.writeResponse(&ErrorStruct500{Message: "error parsing form" + err.Error()}, http.StatusInternalServerError)
	}

	expectedBody := new(GmlReqBody)
	err = json.Unmarshal([]byte(r.Form.Get("JSON")), &expectedBody)

	if err != nil {
		myLogger.Printf("processPostMethod: could not unmarshal into the structure we were expecting :: %v", err)
		t.writeResponse(&ErrorStruct500{Message: err.Error()}, http.StatusInternalServerError)
	}

	license, err := getLicenseForDA(expectedBody.Username, expectedBody.Password, expectedBody.RequestID, expectedBody.RequestEncKey)

	if err != nil {
		myLogger.Printf("getLicenseForDA: %v", err)
		t.writeResponse(&ErrorStruct500{Message: err.Error()}, http.StatusInternalServerError)
	}
	respBody := new(GmlResp)
	respBody.Body.License = license
	t.writeResponse(&respBody.Body, http.StatusOK)
}
func (t *GmlServer) uiHandler(w http.ResponseWriter, r *http.Request) {

	t.writer = w

	switch {
	case r.Method == "GET":
		t.processGetMethod()
	case r.Method == "POST":
		t.processPostMethod(r)
	}
}

func (t *GmlServer) writeResponse(response interface{}, code int) error {
	var payload []byte
	var err error
	switch response.(type) {
	case string:
		payload = []byte(response.(string))
	case []byte:
		payload = response.([]byte)
	default:
		payload, err = json.Marshal(response)
		if err != nil {
			return fmt.Errorf("unable to marshal request, reason: %s", err)
		}
	}
	t.writer.WriteHeader(code)
	t.writer.Write(payload)
	return nil

}

func (t *GmlServer) gmlHandler(w http.ResponseWriter, r *http.Request) {

	t.writer = w
	request, err := ioutil.ReadAll(r.Body)
	if err != nil {
		myLogger.Printf("could not read request body :: %v", err)
		t.writeResponse(&ErrorStruct500{Message: err.Error()}, http.StatusInternalServerError)
	}
	expectedBody := new(GmlReqBody)
	err = json.Unmarshal(request, &expectedBody)
	if err != nil {
		myLogger.Printf("handler recieved unexpected body: could not unmarshal into the structure we were expecting :: %v", err)
		t.writeResponse(&ErrorStruct500{Message: err.Error()}, http.StatusInternalServerError)
	}

	license, err := getLicenseForDA(expectedBody.Username, expectedBody.Password, expectedBody.RequestID, expectedBody.RequestEncKey)
	if err != nil {
		myLogger.Printf("getLicenseForDA: %v", err)
		t.writeResponse(&ErrorStruct500{Message: err.Error()}, http.StatusInternalServerError)
	}
	respBody := new(GmlResp)
	respBody.Body.License = license
	t.writeResponse(&respBody.Body, http.StatusOK)

}

func (t *GmlServer) Start() (server *http.Server, err error) {
	server = &http.Server{Addr: t.Port}

	http.HandleFunc("/"+t.UIPath, t.uiHandler)
	http.HandleFunc("/gml", t.gmlHandler)
	err = t.startServer(server)
	if err != nil {
		myLogger.Printf("Error starting server: %s", err)
		return nil, err
	}
	return server, nil
}

func (t *GmlServer) startServer(server *http.Server) error {
	var certFile string
	var keyFile string
	var err error
	if viper.GetBool("http.tls.enabled") {
		certFile = viper.GetString("http.tls.cert.file")
		keyFile = viper.GetString("http.tls.key.file")
		err = server.ListenAndServeTLS(certFile, keyFile)
	} else {
		//use your own listener so we can close the server when we want!
		t.listener, err = net.Listen("tcp", fmt.Sprintf("%s", t.Port))
		if err != nil {
			myLogger.Printf("%s\n", t.ServerAddress)
			myLogger.Fatalf("startup error: %v", err.Error())
		}
		err = server.Serve(t.listener)
	}
	return err
}

func (t *GmlServer) Close() error {
	if t.listener == nil {
		return nil
	}
	err := t.listener.Close()
	if err != nil {
		return err
	}
	return nil
}

func (t *GmlServer) initConfig(cfgFile string) error {
	err := config.SetupViper(cfgFile)
	if err != nil {
		return fmt.Errorf("failed to set up viper using config file and environmental variables %v", err)
	}

	t.ServerAddress = viper.GetString(SERVER_ADDRESS)
	t.Port = viper.GetString(SERVER_PORT)
	t.UIPath = viper.GetString(SERVER_UI_PATH)
	t.SimServerURL = viper.GetString(SIMSERVER_URL)
	t.MyamURL = viper.GetString(MYAM_URL)

	Config.CorrectProviderURL = t.MyamURL + "/myam/oidc"
	Config.CorrectAudience = t.MyamURL + "/myam/oidc/token"
	Config.MyBankBaseURL = t.SimServerURL + "/my-bank"
	Config.UILocales = "en"
	Config.SimServerURL = t.SimServerURL

	myLogger.Printf("simulator Web UI is up on: " + t.ServerAddress + "/" + t.UIPath)
	myLogger.Printf("config initialization has completed.")
	return nil
}
