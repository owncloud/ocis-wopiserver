package svc

import (
	"context"
	"crypto/tls"
	"encoding/base64"
	"encoding/json"
	"errors"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"path/filepath"
	"strings"
	"time"
	"unicode/utf8"

	merrors "github.com/asim/go-micro/v3/errors"
	gateway "github.com/cs3org/go-cs3apis/cs3/gateway/v1beta1"
	rpc "github.com/cs3org/go-cs3apis/cs3/rpc/v1beta1"
	provider "github.com/cs3org/go-cs3apis/cs3/storage/provider/v1beta1"
	"github.com/cs3org/reva/pkg/auth/scope"
	"github.com/cs3org/reva/pkg/token"
	revajwt "github.com/cs3org/reva/pkg/token/manager/jwt"
	revauser "github.com/cs3org/reva/pkg/user"
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
		httpClient: &http.Client{Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: options.Config.WopiServer.Insecure,
			},
		}},
		client: options.CS3Client,
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
	serviceID  string
	logger     log.Logger
	config     *config.Config
	mux        *chi.Mux
	httpClient *http.Client
	client     gateway.GatewayAPIClient
}

// ServeHTTP implements the Service interface.
func (p WopiServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	p.mux.ServeHTTP(w, r)
}

func (p WopiServer) NotFound(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotFound)
}

type WopiResponse struct {
	WopiClientURL  string `json:"wopiclienturl"`
	AccessToken    string `json:"accesstoken"`
	AccessTokenTTL int64  `json:"accesstokenttl"`
}

func (p WopiServer) OpenFile(w http.ResponseWriter, r *http.Request) {

	username, revaToken, err := getUserAndAuthToken(r, p.config.TokenManager)
	if err != nil {
		p.logger.Err(err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	fileID := r.URL.Query().Get("fileId")
	if fileID == "" {
		pathErr := errors.New("fileID parameter missing in request")
		p.logger.Err(pathErr)
		http.Error(w, pathErr.Error(), http.StatusBadRequest)
		return
	}

	statResponse, err := p.stat(fileID, revaToken)
	if err != nil {
		p.logger.Err(err)
		http.Error(w, "could not stat file", http.StatusBadRequest)
		return
	}

	extensions, err := p.getExtensions()
	if err != nil {
		p.logger.Err(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	extensionHandler, found := extensions[filepath.Ext(statResponse.Info.Path)]
	if !found {
		err = errors.New("file type " + filepath.Ext(statResponse.Info.Path) + " is not supported")
		p.logger.Err(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	wopiClientHost := ""
	viewMode := ""
	canEdit := statResponse.Info.PermissionSet.InitiateFileUpload
	canView := statResponse.Info.PermissionSet.InitiateFileDownload
	isEmpty := statResponse.Info.Size == 0

	if canEdit && canView && isEmpty {
		wopiClientHost = extensionHandler.NewURL //let WOPI client do the file initialization
		viewMode = "VIEW_MODE_READ_WRITE"
	} else if canEdit && canView && !isEmpty {
		wopiClientHost = extensionHandler.EditURL
		viewMode = "VIEW_MODE_READ_WRITE"
	} else if !canEdit && canView && !isEmpty {
		wopiClientHost = extensionHandler.ViewURL
		viewMode = "VIEW_MODE_READ_ONLY"
		//} else if !canEdit && canView && !isEmpty {
		//	 TODO: this branch will never be entered
		//	 permission set is not really useful for this case -> need to use this https://github.com/cs3org/cs3apis/blob/master/cs3/app/provider/v1beta1/provider_api.proto#L79
		//	wopiClientHost = extensionHandler.ViewURL
		//	viewMode = "VIEW_MODE_VIEW_ONLY"
	} else {
		return
	}

	wopiSrc, err := p.getWopiSrc(
		statResponse.Info.Id.OpaqueId, viewMode,
		statResponse.Info.Id.StorageId, filepath.Dir(statResponse.Info.Path),
		username, revaToken,
	)
	if err != nil {
		p.logger.Err(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	u, err := url.Parse(wopiClientHost + "&WOPISrc=" + wopiSrc)
	if err != nil {
		p.logger.Err(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	q := u.Query()

	// remove access token from query parameters
	accessToken := q.Get("access_token")
	q.Del("access_token")

	// more options used by oC 10:
	// &lang=en-GB
	// &closebutton=1
	// &revisionhistory=1
	// &title=Hello.odt
	u.RawQuery = q.Encode()

	js, err := json.Marshal(
		WopiResponse{
			WopiClientURL: u.String(),
			AccessToken:   accessToken,
			// https://wopi.readthedocs.io/projects/wopirest/en/latest/concepts.html#term-access-token-ttl
			AccessTokenTTL: time.Now().Add(p.config.TokenManager.TokenTTL).UnixNano() / 1e6,
		},
	)
	if err != nil {
		p.logger.Err(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
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

	r, err := p.httpClient.Get(p.config.WopiServer.Host + "/wopi/cbox/endpoints")
	if err != nil {
		return nil, err
	}
	defer r.Body.Close()

	if r.StatusCode != http.StatusOK {
		return nil, errors.New("get /wopi/cbox/endpoints failed: status code != 200")
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		return nil, err
	}

	extensions = map[string]ExtensionHandler{}
	err = json.Unmarshal(body, &extensions)
	if err != nil {
		return nil, err
	}

	return extensions, err
}

func (p WopiServer) getWopiSrc(fileRef, viewMode, storageID, folderURL, userName, revaToken string) (b string, err error) {

	req, err := http.NewRequest("GET", p.config.WopiServer.Host+"/wopi/iop/open", nil)
	if err != nil {
		return "", err
	}

	req.Header.Add("authorization", "Bearer "+p.config.WopiServer.IOPSecret)
	req.Header.Add("TokenHeader", revaToken)

	q := req.URL.Query()
	q.Add("filename", fileRef) // can be the file path or an opaque ID
	q.Add("viewmode", viewMode)
	q.Add("folderurl", folderURL)
	q.Add("endpoint", storageID)
	q.Add("username", userName)
	req.URL.RawQuery = q.Encode()

	r, err := p.httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer r.Body.Close()

	if r.StatusCode != http.StatusOK {
		return "", errors.New("get /wopi/iop/open failed: status code != 200")
	}

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return "", err
	}

	return string(body), err
}

func (p WopiServer) stat(fileID, auth string) (*provider.StatResponse, error) {
	ctx := metadata.AppendToOutgoingContext(context.Background(), token.TokenHeader, auth)

	// taken from reva - ocdav
	unwrap := func(rid string) *provider.ResourceId {
		decodedID, err := base64.URLEncoding.DecodeString(rid)
		if err != nil {
			return nil
		}

		parts := strings.SplitN(string(decodedID), ":", 2)
		if len(parts) != 2 {
			return nil
		}

		if !utf8.ValidString(parts[0]) || !utf8.ValidString(parts[1]) {
			return nil
		}

		return &provider.ResourceId{
			StorageId: parts[0],
			OpaqueId:  parts[1],
		}
	}

	resourceID := unwrap(fileID)
	if resourceID == nil {
		return nil, errors.New("unwrap fileID failed")
	}

	req := &provider.StatRequest{
		Ref: &provider.Reference{
			Spec: &provider.Reference_Id{
				Id: resourceID,
			},
		},
	}
	rsp, err := p.client.Stat(ctx, req)
	if err != nil {
		p.logger.Error().Err(err).Str("fileID", fileID).Msg("could not stat file")
		return nil, merrors.InternalServerError(p.serviceID, "could not stat file: %s", err.Error())
	}

	if rsp.Status.Code != rpc.Code_CODE_OK {
		switch rsp.Status.Code {
		case rpc.Code_CODE_NOT_FOUND:
			return nil, merrors.NotFound(p.serviceID, "could not stat file: %s", rsp.Status.Message)
		default:
			p.logger.Error().Str("status_message", rsp.Status.Message).Str("fileID", fileID).Msg("could not stat file")
			return nil, merrors.InternalServerError(p.serviceID, "could not stat file: %s", rsp.Status.Message)
		}
	}
	if rsp.Info.Type != provider.ResourceType_RESOURCE_TYPE_FILE {
		return nil, merrors.BadRequest(p.serviceID, "Unsupported file type")
	}
	return rsp, nil
}

func getUserAndAuthToken(r *http.Request, tm config.TokenManager) (username, revaToken string, err error) {

	ctx := r.Context()

	tokenManager, err := revajwt.New(map[string]interface{}{
		"secret":  tm.JWTSecret,
		"expires": tm.TokenTTL.Seconds(),
	})
	if err != nil {
		return "", "", err
	}

	user := revauser.ContextMustGetUser(ctx)
	scope, err := scope.GetOwnerScope()
	if err != nil {
		return "", "", err
	}

	revaToken, err = tokenManager.MintToken(ctx, user, scope)
	if err != nil {
		return "", "", err
	}

	username = user.DisplayName

	// TODO: if CS3org WOPI server mints the final REVA JWT secret,
	// the temporary REVA JWT token and the user display name can also be obtained like that:
	//
	//revaToken = r.Header.Get("X-Access-Token") // reva token minted by oCIS Proxy
	//if revaToken == "" {
	//	return "", "", errors.New("unauthenticated request")
	//}
	//
	//type revaClaims struct {
	//	User *user.User `json:"user,omitempty"`
	//	jwt.Claims
	//}
	//
	//var claims revaClaims
	//
	////decode JWT token without verifying the signature
	//tokenErr := errors.New("request provided malformed access token")
	//token, err := jwt.ParseSigned(revaToken)
	//if err != nil {
	//	return "", "", tokenErr
	//}
	//err = token.UnsafeClaimsWithoutVerification(&claims)
	//if err != nil {
	//	return "", "", tokenErr
	//}
	//
	//username = claims.User.DisplayName

	return username, revaToken, nil
}
