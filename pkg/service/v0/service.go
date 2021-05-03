package svc

import (
	"crypto/tls"
	"encoding/json"
	"errors"
	"net/http"
	"path/filepath"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/owncloud/ocis-wopiserver/pkg/assets"
	"github.com/owncloud/ocis-wopiserver/pkg/config"
	"github.com/owncloud/ocis/ocis-pkg/log"
	ocsm "github.com/owncloud/ocis/ocis-pkg/middleware"
)

// Service defines the extension handlers.
type Service interface {
	ServeHTTP(http.ResponseWriter, *http.Request)
}

// NewService returns a service implementation for Service.
func NewService(opts ...Option) Service {
	options := newOptions(opts...)

	m := chi.NewMux()
	m.Use(options.Middleware...)

	m.Use(ocsm.Static(
		options.Config.HTTP.Root,
		assets.New(
			assets.Logger(options.Logger),
			assets.Config(options.Config),
		),
		options.Config.HTTP.CacheTTL,
	))

	svc := WopiServer{
		logger: options.Logger,
		config: options.Config,
		mux:    m,
	}

	m.Route(options.Config.HTTP.Root, func(r chi.Router) {
		r.NotFound(svc.NotFound)
		r.Use(middleware.StripSlashes)
		r.Get("/api/v0/wopi/open", svc.OpenFile)

	})

	return svc
}

// WopiServer defines implements the business logic for Service.
type WopiServer struct {
	logger log.Logger
	config *config.Config
	mux    *chi.Mux
}

// ServeHTTP implements the Service interface.
func (p WopiServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	p.mux.ServeHTTP(w, r)
}

func (p WopiServer) NotFound(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotFound)
}

type WopiResponse struct {
	WopiAccessToken string `json:"accesstoken"`
	WopiClientURL   string `json:"wopiclienturl"`
}

type WopiToken struct {
	ViewMode  string `json:"viewmode"`
	Token     string `json:"userid"`
	StorageID string `json:"endpoint"`
	FolderURL string `json:"folderurl"`
	FilePath  string `json:"filename"`
	UserName  string `json:"username"`
	jwt.StandardClaims
}

func (p WopiServer) OpenFile(w http.ResponseWriter, r *http.Request) {
	filePath := r.URL.Query().Get("filePath")
	if filePath == "" {
		w.WriteHeader(http.StatusBadRequest)
		p.logger.Err(errors.New("filePath parameter missing in request"))
		return
	}

	token := r.Header.Get("X-Access-Token")
	if token == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	mode := "write" // TODO: decide how to open it

	viewmode := ""

	switch mode {
	case "view":
		viewmode = "VIEW_MODE_VIEW_ONLY"
	case "read":
		viewmode = "VIEW_MODE_READ_ONLY"
	case "write":
		viewmode = "VIEW_MODE_READ_WRITE"
	case "default":
		return
	}

	wt := WopiToken{
		ViewMode:  viewmode,
		Token:     token,
		StorageID: "1284d238-aa92-42ce-bdc4-0b0000009157", // TODO: do not hardcode
		FolderURL: "",
		FilePath:  filePath,
		UserName:  "Einstein", // TODO: get user from context
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Unix() + (60 * 60), // TODO: decide about expiry
		},
	}

	swt := jwt.NewWithClaims(jwt.SigningMethodHS256, wt)
	wopiToken, err := swt.SignedString([]byte(p.config.WopiServer.WopiServerSecret))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		p.logger.Err(err)
		return
	}

	extensions, err := getExtensions(p.config.WopiServer.WopiServerHost, p.config.WopiServer.WopiServerInsecure)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		p.logger.Err(err)
		return
	}

	extensionHandler, found := extensions[filepath.Ext(filePath)]

	if !found {
		err = errors.New("file type " + filepath.Ext(filePath) + " is not supported")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		p.logger.Err(err)
		return
	}

	wopiClientHost := ""

	switch mode {
	case "view":
		wopiClientHost = extensionHandler.ViewURL
	case "read":
		wopiClientHost = extensionHandler.ViewURL
	case "write":
		wopiClientHost = extensionHandler.EditURL
	case "default":
		return
	}

	res := WopiResponse{
		WopiAccessToken: wopiToken,
		WopiClientURL:   wopiClientHost + "?WOPISrc=" + p.config.WopiServer.WopiServerHost + "/wopi/files/1", // TODO: set URI even if totally unused
	}

	js, err := json.Marshal(res)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		p.logger.Err(err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(js)
}

type ExtensionHandler struct {
	ViewURL string `json:"view"`
	EditURL string `json:"edit"`
	NewURL  string `json:"new"`
}

func getExtensions(wopiServerHost string, insecure bool) (extensions map[string]ExtensionHandler, err error) {

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: insecure},
	}
	c := &http.Client{Transport: tr}

	r, err := c.Get(wopiServerHost + "/wopi/cbox/endpoints")
	if err != nil {
		return nil, err
	}
	defer r.Body.Close()

	extensions = map[string]ExtensionHandler{}

	err = json.NewDecoder(r.Body).Decode(&extensions)
	return extensions, err
}
