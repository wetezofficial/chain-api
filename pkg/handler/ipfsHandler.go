package handler

import (
	"context"
	"fmt"
	"github.com/ipfs-cluster/ipfs-cluster/api/rest/client"
	files "github.com/ipfs/go-ipfs-files"
	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
	"io"
	"mime/multipart"
	"starnet/chain-api/pkg/app"
	"starnet/chain-api/pkg/appcontext"
	"starnet/starnet/constant"
	"starnet/starnet/models"
	"time"
)

type IPFSCluster struct {
	IPFSClient  client.Client
	JsonHandler *JsonRpcHandler
}

func NewIPFSCluster(
	chain constant.Chain,
	app *app.App,
) *IPFSCluster {
	return &IPFSCluster{
		JsonHandler: NewJsonRpcHandler(chain, nil, nil, nil, nil, app),
	}
}

func (h *IPFSCluster) newLogger(c echo.Context) *zap.Logger {
	// add chain name
	return h.JsonHandler.logger.With(zap.String("chain", h.JsonHandler.chain.Name), zap.String("request_id", c.Request().Context().Value("request_id").(string)))
}

func (h *IPFSCluster) Upload(c echo.Context) error {
	logger := h.newLogger(c)

	apiKey, err := h.JsonHandler.bindApiKey(c)
	if err != nil {
		return err
	}

	if rlErr := h.JsonHandler.rateLimit(c.Request().Context(), logger, apiKey, 1); rlErr != nil {
		logger.Debug("rate limit", zap.String("apiKey", apiKey), zap.Error(rlErr))
		return c.JSON(200, rlErr)
	}

	// TODO: proxy request to ipfs cluster
	// get cid from response
	// list cid files
	// save the files to database

	//ctx, cancelFunc := context.WithTimeout(c.Request().Context(), time.Second*5)
	//defer cancelFunc()

	// TODO: Upload
	cc := c.(*appcontext.AppContext)

	// Multipart form
	form, err := c.MultipartForm()
	if err != nil {
		return c.JSON(400, err)
	}
	requestFiles := form.File["files"]
	if len(requestFiles) == 0 {
		return c.JSON(200, nil)
	} else if len(requestFiles) == 1 {
		file, err := requestFiles[0].Open()
		if err != nil {
			return c.JSON(400, err)
		}
		defer func(file multipart.File) {
			_ = file.Close()
		}(file)
		cid, err := cc.IPFSSrv.Upload(c.Request().Context(), io.Reader(file))
		if err != nil {
			return err
		}
		return c.JSON(200, cid)
	}

	fs := make(map[string]files.Node)

	sf := files.NewMapDirectory(fs)

	//sf := NewMapDirectory(map[string]Node{
	//	"file.txt": NewBytesFile([]byte(text)),
	//	"boop": NewMapDirectory(map[string]Node{
	//		"a.txt": NewBytesFile([]byte("bleep")),
	//		"b.txt": NewBytesFile([]byte("bloop")),
	//	}),
	//	"beep.txt": NewBytesFile([]byte("beep")),
	//})

	cid, err := cc.IPFSSrv.UploadDir(c.Request().Context(), files.NewMultiFileReader(sf, true))
	if err != nil {
		return c.JSON(400, err)
	}

	return c.JSON(200, cid)
}

func (h *IPFSCluster) List(c echo.Context) error {
	logger := h.newLogger(c)

	apiKey, err := h.JsonHandler.bindApiKey(c)
	if err != nil {
		return err
	}

	if rlErr := h.JsonHandler.rateLimit(c.Request().Context(), logger, apiKey, 1); rlErr != nil {
		logger.Debug("rate limit", zap.String("apiKey", apiKey), zap.Error(rlErr))
		return c.JSON(200, rlErr)
	}

	ctx, cancelFunc := context.WithTimeout(c.Request().Context(), time.Second*5)
	defer cancelFunc()

	cc := c.(*appcontext.AppContext)
	var fileList []models.IPFSFile
	// TODO: apikey to query userID
	err = cc.IPFSSrv.ListUserFile(ctx, 0, &fileList)
	if err != nil {
		return c.JSON(400, err)
	}
	return c.JSON(200, fileList)
}

func (h *IPFSCluster) Get(c echo.Context) error {
	logger := h.newLogger(c)

	apiKey, err := h.JsonHandler.bindApiKey(c)
	if err != nil {
		return err
	}

	if rlErr := h.JsonHandler.rateLimit(c.Request().Context(), logger, apiKey, 1); rlErr != nil {
		logger.Debug("rate limit", zap.String("apiKey", apiKey), zap.Error(rlErr))
		return c.JSON(200, rlErr)
	}

	ctx, cancelFunc := context.WithTimeout(c.Request().Context(), time.Second*5)
	defer cancelFunc()

	// TODO: Get
	// proxy request to ipfs gateway
	fmt.Println(ctx.Err())

	// TODO: add file size compute
	return c.JSONBlob(200, nil)
}

func (h *IPFSCluster) Pin(c echo.Context) error {
	logger := h.newLogger(c)

	apiKey, err := h.JsonHandler.bindApiKey(c)
	if err != nil {
		return err
	}

	if rlErr := h.JsonHandler.rateLimit(c.Request().Context(), logger, apiKey, 1); rlErr != nil {
		logger.Debug("rate limit", zap.String("apiKey", apiKey), zap.Error(rlErr))
		return c.JSON(200, rlErr)
	}

	ctx, cancelFunc := context.WithTimeout(c.Request().Context(), time.Second*5)
	defer cancelFunc()

	// TODO: Pin
	fmt.Println(ctx.Err())

	// FIXME: resp cid
	return c.JSON(200, nil)
}
