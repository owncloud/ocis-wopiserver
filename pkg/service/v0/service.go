package svc

import (
	"crypto/tls"
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"path/filepath"

	"github.com/cs3org/reva/pkg/auth/scope"
	"github.com/cs3org/reva/pkg/token/manager/jwt"
	"github.com/cs3org/reva/pkg/user"
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
		c: &http.Client{Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: options.Config.WopiServer.Insecure,
			},
		}},
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
	c      *http.Client
}

// ServeHTTP implements the Service interface.
func (p WopiServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	p.mux.ServeHTTP(w, r)
}

func (p WopiServer) NotFound(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotFound)
}

type WopiResponse struct {
	WopiClientURL string `json:"wopiclienturl"`
}

func (p WopiServer) OpenFile(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	filePath := r.URL.Query().Get("filePath")
	if filePath == "" {
		w.WriteHeader(http.StatusBadRequest)
		p.logger.Err(errors.New("filePath parameter missing in request"))
		return
	}

	folderPath, _ := filepath.Split(filePath)

	tokenManager, err := jwt.New(map[string]interface{}{
		"secret":  p.config.TokenManager.JWTSecret,
		"expires": int64(60),
	})
	if err != nil {
		p.logger.Err(err)
		return
	}

	user := user.ContextMustGetUser(ctx)
	scope, err := scope.GetOwnerScope()
	if err != nil {
		p.logger.Err(err)
		return
	}

	revaToken, err := tokenManager.MintToken(ctx, user, scope)
	if err != nil {
		p.logger.Err(err)
		return
	}

	mode := "write"                                     // TODO: decide how to open it
	storageID := "1284d238-aa92-42ce-bdc4-0b0000009157" // TODO: make dynamic

	extensions, err := p.getExtensions()
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
	viewMode := ""

	switch mode {
	case "view":
		wopiClientHost = extensionHandler.ViewURL
		viewMode = "VIEW_MODE_VIEW_ONLY"
	case "read":
		wopiClientHost = extensionHandler.ViewURL
		viewMode = "VIEW_MODE_READ_ONLY"
	case "write":
		wopiClientHost = extensionHandler.EditURL
		viewMode = "VIEW_MODE_READ_WRITE"
	case "default":
		return
	}

	wopiSrc, err := p.getWopiSrc(filePath, viewMode, storageID, folderPath, revaToken)
	if err != nil {
		p.logger.Err(err)
		return
	}

	// more options used by oC 10
	// &lang=en-GB
	// &closebutton=1
	// &revisionhistory=1
	// &title=Hello.odt
	res := WopiResponse{
		WopiClientURL: wopiClientHost + "?WOPISrc=" + wopiSrc,
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

func (p WopiServer) getExtensions() (extensions map[string]ExtensionHandler, err error) {

	r, err := p.c.Get(p.config.WopiServer.Host + "/wopi/cbox/endpoints")
	if err != nil {
		return nil, err
	}
	defer r.Body.Close()

	if r.StatusCode != http.StatusOK {
		return nil, errors.New("get /wopi/cbox/endpoints failed: status code != 200")
	}

	extensions = map[string]ExtensionHandler{}

	err = json.NewDecoder(r.Body).Decode(&extensions)
	return extensions, err
}

func (p WopiServer) getWopiSrc(filePath string, viewMode string, storageID string, folderURL string, revaToken string) (resp string, err error) {

	req, err := http.NewRequest("GET", p.config.WopiServer.Host+"/wopi/iop/open", nil)

	req.Header.Add("authorization", "Bearer "+p.config.WopiServer.Secret)
	req.Header.Add("TokenHeader", revaToken)

	q := req.URL.Query()
	q.Add("filename", filePath)
	q.Add("viewmode", viewMode)
	q.Add("folderurl", folderURL)
	q.Add("endpoint", storageID)
	req.URL.RawQuery = q.Encode()

	r, err := p.c.Do(req)
	if err != nil {
		return "", err
	}
	defer r.Body.Close()

	if r.StatusCode != http.StatusOK {
		return "", errors.New("get /wopi/iop/open failed: status code != 200")
	}

	b, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return "", err
	}
	return string(b), err
}
