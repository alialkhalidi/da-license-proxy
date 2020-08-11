package gmlserver

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	yaml "gopkg.in/yaml.v2"

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
}

func NewGmlServer(cfgFile string) (*GmlServer, error) {
	instance := GmlServer{}
	err := instance.initConfig(cfgFile)

	return &instance, err

}

func getLicenseForDA(username, password, licenseRequestID, requestEncKey string) (string, error) {

	accessToken, err := GetAccessToken(VerifiedMeScope, username, password, "")
	if err != nil {
		myLogger.Printf("getLicenseForDA->GetAccessToken: %v", err)
		return "", err
	}
	serverState, _, err := RecoverLockboxWithClientID(accessToken, http.StatusAccepted, "")
	if err != nil {
		myLogger.Printf("getLicenseForDA->RecoverLockboxWithClientID for user %s: %v", username, err)
		return "", err
	}

	assets := []string{"vme://assets/foundationalIdentity"}
	//tempServerState := serverState
	serverState, daMap, err := CreateDA(accessToken, serverState, assets)
	/*
		if strings.Compare(tempServerState, serverState) == 0 {
			myLogger.Printf("getLicenseForDA->CreateDA for user %s: no server state chaneg containsExactAssets", username)
		}
	*/
	if err != nil {
		myLogger.Printf("getLicenseForDA->CreateDA for user %s: %v", username, err)
		return "", err
	}

	serverState, _, err = RetrieveLicenseRequest(accessToken, serverState, licenseRequestID, requestEncKey, http.StatusAccepted)
	if err != nil {
		myLogger.Printf("getLicenseForDA->RetrieveLicenseRequest for user %s: %v", username, err)
		return "", err
	}

	issueLicenseResp, err := IssueLicense(accessToken, serverState, licenseRequestID, daMap)
	if err != nil {
		myLogger.Printf("getLicenseForDA->IssueLicense for user %s: %v", username, err)
		return "", err
	}
	return issueLicenseResp.Body.License, nil

}

func (t *GmlServer) processGetMethod(w http.ResponseWriter) {
	const page = `<html>
  <form id="gml" action="/ui" method="post">
  <textarea name="JSON" id="JSON" placeholder='{"username": "", "password": "", "requestId": "", "requestEncKey": ""}' spellcheck="false" rows="20" form="gml"></textarea>
  <input type="submit" value="Send Request<"/>
  </form>
  <html>
  `
	fmt.Fprintf(w, "%s", page)
}

func (t *GmlServer) processPostMethod(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		myLogger.Printf("processPostMethod: error parsing form")
		t.writeResponse(w, &ErrorStruct500{Message: "error parsing form" + err.Error()}, http.StatusInternalServerError)
		return
	}

	expectedBody := new(GmlReqBody)
	err = json.Unmarshal([]byte(r.Form.Get("JSON")), &expectedBody)
	if err != nil {
		myLogger.Printf("processPostMethod: could not unmarshal into the structure we were expecting :: %v", err)
		t.writeResponse(w, &ErrorStruct500{Message: err.Error()}, http.StatusInternalServerError)
		return
	}

	license, err := getLicenseForDA(expectedBody.Username, expectedBody.Password, expectedBody.RequestID, expectedBody.RequestEncKey)

	if err != nil {
		myLogger.Printf("processPostMethod->getLicenseForDA : %v", err)
		t.writeResponse(w, &ErrorStruct500{Message: err.Error()}, http.StatusInternalServerError)
		return
	}
	respBody := new(GmlResp)
	respBody.Body.License = license
	t.writeResponse(w, &respBody.Body, http.StatusOK)
}
func (t *GmlServer) uiHandler(w http.ResponseWriter, r *http.Request) {
	switch {
	case r.Method == "GET":
		t.processGetMethod(w)
	case r.Method == "POST":
		t.processPostMethod(w, r)
	}
}

func (t *GmlServer) writeResponse(w http.ResponseWriter, response interface{}, code int) error {
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
	w.WriteHeader(code)
	w.Write(payload)
	return nil

}

func (t *GmlServer) gmlHandler(w http.ResponseWriter, r *http.Request) {

	request, err := ioutil.ReadAll(r.Body)
	if err != nil {
		myLogger.Printf("could not read request body :: %v", err)
		t.writeResponse(w, &ErrorStruct500{Message: err.Error()}, http.StatusInternalServerError)
		return
	}
	expectedBody := new(GmlReqBody)
	err = json.Unmarshal(request, &expectedBody)
	if err != nil {
		myLogger.Printf("handler recieved unexpected body: could not unmarshal into the structure we were expecting :: %v", err)
		t.writeResponse(w, &ErrorStruct500{Message: err.Error()}, http.StatusInternalServerError)
		return
	}

	license, err := getLicenseForDA(expectedBody.Username, expectedBody.Password, expectedBody.RequestID, expectedBody.RequestEncKey)
	if err != nil {
		myLogger.Printf("getLicenseForDA: %v", err)
		t.writeResponse(w, &ErrorStruct500{Message: err.Error()}, http.StatusInternalServerError)
		return
	}
	respBody := new(GmlResp)
	respBody.Body.License = license
	t.writeResponse(w, &respBody.Body, http.StatusOK)

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

func (t *GmlServer) setupViper(cfgFile string) error {
	v := viper.GetViper()
	var err error
	var data []byte
	confType := "yaml"

	data, err = ioutil.ReadFile(filepath.Clean(cfgFile))
	if err != nil {
		return fmt.Errorf("failed to read configuration file %s. %v", cfgFile, err)
	}

	ext := filepath.Ext(cfgFile)
	if len(ext) > 1 {
		confType = ext[1:len(ext)]
	}

	cfgStr := string(data)
	envExpandedCfg := os.ExpandEnv(cfgStr)
	v.SetConfigType(confType)
	v.AutomaticEnv()
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_", "-", "_"))

	err = v.ReadConfig(bytes.NewReader([]byte(envExpandedCfg)))
	if err != nil {
		return fmt.Errorf("failed to load configuration data. %v", err)
	}

	yamlMap := make(map[string]interface{})
	err = v.Unmarshal(&yamlMap)
	if err != nil {
		return fmt.Errorf("viper.Unmarshal failed %v", err)
	}

	rawCfg, err := yaml.Marshal(&yamlMap)
	if err != nil {
		return fmt.Errorf("failed to marshal configuration content into YAML %v", err)
	}

	err = v.ReadConfig(bytes.NewReader(rawCfg))
	if err != nil {
		return fmt.Errorf("readconfig failed: %v", err)
	}
	return nil

}

func (t *GmlServer) initConfig(cfgFile string) error {
	err := t.setupViper(cfgFile)
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
