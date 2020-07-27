package gmlserver

import (
	"fmt"
	"log"
	"net"
	"net/http"

	"sk-git.securekey.com/gerrit/vme-core/utils/config"

	"github.com/spf13/viper"
)

var myLogger = log.New(log.Writer(), "gmlserver ", 0)

const (
	clientAuthBasic = "basic"

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

func uiHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "UI All good\n")
}

func GmlHandler(w http.ResponseWriter, r *http.Request) {

	accessToken, err := GetAccessToken(VerifiedMeScope, "jdoe7", "password", "")
	if err != nil {
		myLogger.Fatalf("GetAccessToken: %v", err)
	}
	fmt.Fprintf(w, "GetAccessToken: Success\n")
	serverState, _, err := RecoverLockboxWithClientID(accessToken, http.StatusAccepted, "")
	if err != nil {
		myLogger.Fatalf("RecoverLockboxWithClientID: %v", err)
	}
	fmt.Fprintf(w, "RecoverLockboxWithClientID: Success\n")

	assets := []string{"vme://assets/foundationalIdentity"}
	serverState, daMap, err := CreateDA(accessToken, serverState, assets)
	if err != nil {
		myLogger.Fatalf("CreateDA: %v", err)
	}
	fmt.Fprintf(w, "CreateDA: Success\n")

	licenseRequestID := "_l3xIhxOWpvJSqVRQCskC5HJVZhD9KoaUr5FK7z545I"
	requestEncKey := "AV6V5gZMQwc0dksDEoHNAhIX-AckTXpgDZOVYtserV0"
	serverState, _, err = RetrieveLicenseRequest(accessToken, serverState, licenseRequestID, requestEncKey, http.StatusAccepted)
	if err != nil {
		myLogger.Fatalf("RetrieveLicenseRequest: %v", err)
	}
	fmt.Fprintf(w, "RetrieveLicenseRequest: Success\n")
	issueLicenseResp, err := IssueLicense(accessToken, serverState, licenseRequestID, daMap)
	if err != nil {
		myLogger.Fatalf("IssueLicense: %v", err)
	}
	fmt.Fprintf(w, "IssueLicense: Success\n %s", issueLicenseResp.Body.License)
}

func (t *GmlServer) Start() (server *http.Server, err error) {
	server = &http.Server{Addr: t.Port}

	http.HandleFunc("/"+t.UIPath, uiHandler)
	http.HandleFunc("/gml", GmlHandler)
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
