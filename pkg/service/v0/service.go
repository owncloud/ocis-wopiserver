package svc

import (
	"context"
	"crypto/tls"
	"encoding/base64"
	"encoding/json"
	"errors"
	"net/http"
	"net/url"
	"os/user"
	"strings"
	"unicode/utf8"

	gateway "github.com/cs3org/go-cs3apis/cs3/gateway/v1beta1"
	provider "github.com/cs3org/go-cs3apis/cs3/storage/provider/v1beta1"
	"github.com/cs3org/reva/pkg/token"
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/owncloud/ocis-wopiserver/pkg/assets"
	"github.com/owncloud/ocis-wopiserver/pkg/config"
	"github.com/owncloud/ocis/ocis-pkg/log"
	ocsm "github.com/owncloud/ocis/ocis-pkg/middleware"
	"google.golang.org/grpc/metadata"
	"gopkg.in/square/go-jose.v2/jwt"
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

	revaToken, err := getToken(r)
	if err != nil {
		p.logger.Logger.Err(err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	fileID := r.URL.Query().Get("fileId")
	if fileID == "" {
		pathErr := errors.New("fileID parameter missing in request")
		p.logger.Logger.Err(pathErr)
		http.Error(w, pathErr.Error(), http.StatusBadRequest)
		return
	}

	ctx := metadata.AppendToOutgoingContext(context.Background(), token.TokenHeader, revaToken)

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
		err := errors.New("unwrap fileID failed")
		p.logger.Logger.Err(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	appOpenReq := gateway.OpenInAppRequest{
		Ref: &provider.Reference{
			ResourceId: resourceID,
		},
		ViewMode: gateway.OpenInAppRequest_VIEW_MODE_READ_WRITE, // TODO: make configurable
	}

	appResp, err := p.client.OpenInApp(ctx, &appOpenReq)
	if err != nil {
		p.logger.Logger.Err(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	u, err := url.Parse(appResp.AppUrl)
	if err != nil {
		p.logger.Logger.Err(err)
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

	wr := WopiResponse{
		WopiClientURL: u.String(),
		AccessToken:   accessToken,
		// https://wopi.readthedocs.io/projects/wopirest/en/latest/concepts.html#term-access-token-ttl
		//AccessTokenTTL: time.Now().Add(p.config.TokenManager.TokenTTL).UnixNano() / 1e6,
	}

	js, err := json.Marshal(
		wr,
	)
	if err != nil {
		p.logger.Logger.Err(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(js)

}

func getToken(r *http.Request) (revaToken string, err error) {

	revaToken = r.Header.Get("X-Access-Token") // reva token minted by oCIS Proxy
	if revaToken == "" {
		return "", errors.New("unauthenticated request")
	}

	type revaClaims struct {
		User *user.User `json:"user,omitempty"`
		jwt.Claims
	}

	var claims revaClaims

	//decode JWT token without verifying the signature
	tokenErr := errors.New("request provided malformed access token")
	token, err := jwt.ParseSigned(revaToken)
	if err != nil {
		return "", tokenErr
	}
	err = token.UnsafeClaimsWithoutVerification(&claims)
	if err != nil {
		return "", tokenErr
	}

	return revaToken, nil
}
