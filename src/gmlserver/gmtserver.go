package gmlserver

import (
  "net"
  "net/http"
  "fmt"
  "log"

  "sk-git.securekey.com/gerrit/vme-core/utils/config"

  "github.com/spf13/viper"
)

var myLogger = log.New(log.Writer(),"gmlserver ",0)

const (
        clientAuthBasic = "basic"

        SERVER_ADDRESS  = "http.url"
        SERVER_PORT     = "http.listen.address"
        SERVER_UI_PATH  = "http.ui.path"
        SIMSERVER_URL   = "simserver.url"
)

type GmlServer struct {
        ServerAddress    string
        Port             string
        DACPort          string
        UIPath           string
        SimServerURL     string
        listener         net.Listener
}

func NewGmlServer(cfgFile string) (*GmlServer, error) {
        instance := GmlServer{}
        err := instance.initConfig(cfgFile)

        return &instance, err

}

func (t *GmlServer) StartGo() *http.Server {
        server, _ := t.startH(true)
        return server
}


func (t *GmlServer) Start() error {

        if _, err := t.startH(false); err != nil {
                myLogger.Fatalf("in process startup error: %v", err.Error())
                return err
        }
        return nil
}

func uiHandler(w http.ResponseWriter, r *http.Request) {
   fmt.Fprintf(w, "UI All good\n")
 }

func GmlHandler(w http.ResponseWriter, r *http.Request) {
     fmt.Fprintf(w, "REST All good\n")
}

func (t *GmlServer) startH(async bool) (server *http.Server, err error) {
        server = &http.Server{Addr: t.Port}

        http.HandleFunc("/"+t.UIPath, uiHandler)
        http.HandleFunc("/Gml", GmlHandler)
        if async {
                go t.startServer(server)
        } else {
                err = t.startServer(server)
        }
        return server, err
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
                        myLogger.Printf("%s\n",t.ServerAddress)
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

        myLogger.Printf("simulator Web UI is up on: " + t.ServerAddress + "/" + t.UIPath)
        myLogger.Printf("config initialization has completed.")
        return nil
}
