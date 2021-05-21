package svc

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"path/filepath"

	merrors "github.com/asim/go-micro/v3/errors"
	gateway "github.com/cs3org/go-cs3apis/cs3/gateway/v1beta1"
	rpc "github.com/cs3org/go-cs3apis/cs3/rpc/v1beta1"
	provider "github.com/cs3org/go-cs3apis/cs3/storage/provider/v1beta1"
	"github.com/cs3org/reva/pkg/auth/scope"
	"github.com/cs3org/reva/pkg/token"
	"github.com/cs3org/reva/pkg/token/manager/jwt"
	"github.com/cs3org/reva/pkg/user"
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/owncloud/ocis-wopiserver/pkg/assets"
	"github.com/owncloud/ocis-wopiserver/pkg/config"
	"github.com/owncloud/ocis/ocis-pkg/log"
	ocsm "github.com/owncloud/ocis/ocis-pkg/middleware"
	"google.golang.org/grpc/metadata"
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
		serviceID: options.Config.HTTP.Namespace + "." + options.Config.Server.Name,
		logger:    options.Logger,
		config:    options.Config,
		mux:       m,
		c: &http.Client{Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: options.Config.WopiServer.Insecure,
			},
		}},
		cs3Client: options.CS3Client,
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
	serviceID string
	logger    log.Logger
	config    *config.Config
	mux       *chi.Mux
	c         *http.Client
	cs3Client gateway.GatewayAPIClient
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

	statResponse, err := p.stat(filePath, revaToken)
	if err != nil {
		p.logger.Err(err)
		return
	}

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
	editPerm := statResponse.Info.PermissionSet.InitiateFileUpload
	viewPerm := statResponse.Info.PermissionSet.InitiateFileDownload
	emtpyFile := statResponse.Info.Size == 0

	if editPerm && viewPerm && emtpyFile {
		wopiClientHost = extensionHandler.NewURL //let WOPI client do the file initialization
		viewMode = "VIEW_MODE_READ_WRITE"
	} else if editPerm && viewPerm && !emtpyFile {
		wopiClientHost = extensionHandler.EditURL
		viewMode = "VIEW_MODE_READ_WRITE"
	} else if !editPerm && viewPerm && !emtpyFile {
		wopiClientHost = extensionHandler.ViewURL
		viewMode = "VIEW_MODE_READ_ONLY"
		//} else if !editPerm && viewPerm && !emtpyFile {
		//	 TODO: this branch will never be entered
		//	 permission set is not really useful for this case -> need to use this https://github.com/cs3org/cs3apis/blob/master/cs3/app/provider/v1beta1/provider_api.proto#L79
		//	wopiClientHost = extensionHandler.ViewURL
		//	viewMode = "VIEW_MODE_VIEW_ONLY"
	} else {
		return
	}

	wopiSrc, err := p.getWopiSrc(filePath, viewMode, statResponse.Info.Id.StorageId, folderPath, user.DisplayName, revaToken)
	if err != nil {
		p.logger.Err(err)
		return
	}

	wopiClientURL := wopiClientHost // already includes ?permission=<readonly/edit>
	wopiClientURL += "&WOPISrc=" + wopiSrc
	// more options used by oC 10:
	// &lang=en-GB
	// &closebutton=1
	// &revisionhistory=1
	// &title=Hello.odt

	js, err := json.Marshal(
		WopiResponse{
			WopiClientURL: wopiClientURL,
		},
	)
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

func (p WopiServer) getWopiSrc(filePath string, viewMode string, storageID string, folderURL string, userName string, revaToken string) (resp string, err error) {

	req, err := http.NewRequest("GET", p.config.WopiServer.Host+"/wopi/iop/open", nil)

	req.Header.Add("authorization", "Bearer "+p.config.WopiServer.Secret)
	req.Header.Add("TokenHeader", revaToken)

	q := req.URL.Query()
	q.Add("filename", filePath)
	q.Add("viewmode", viewMode)
	q.Add("folderurl", folderURL)
	q.Add("endpoint", storageID)
	q.Add("username", userName)
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

func (p WopiServer) stat(path, auth string) (*provider.StatResponse, error) {
	ctx := metadata.AppendToOutgoingContext(context.Background(), token.TokenHeader, auth)

	req := &provider.StatRequest{
		Ref: &provider.Reference{
			Spec: &provider.Reference_Path{Path: path},
		},
	}
	rsp, err := p.cs3Client.Stat(ctx, req)
	if err != nil {
		p.logger.Error().Err(err).Str("path", path).Msg("could not stat file")
		return nil, merrors.InternalServerError(p.serviceID, "could not stat file: %s", err.Error())
	}

	if rsp.Status.Code != rpc.Code_CODE_OK {
		switch rsp.Status.Code {
		case rpc.Code_CODE_NOT_FOUND:
			return nil, merrors.NotFound(p.serviceID, "could not stat file: %s", rsp.Status.Message)
		default:
			p.logger.Error().Str("status_message", rsp.Status.Message).Str("path", path).Msg("could not stat file")
			return nil, merrors.InternalServerError(p.serviceID, "could not stat file: %s", rsp.Status.Message)
		}
	}
	if rsp.Info.Type != provider.ResourceType_RESOURCE_TYPE_FILE {
		return nil, merrors.BadRequest(p.serviceID, "Unsupported file type")
	}
	return rsp, nil
}
