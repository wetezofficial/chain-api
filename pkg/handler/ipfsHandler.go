package handler

import (
	"context"
	"fmt"
	files "github.com/ipfs/go-ipfs-files"
	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
	"starnet/chain-api/pkg/app"
	serviceInterface "starnet/chain-api/service/interface"
	"starnet/starnet/constant"
	"starnet/starnet/models"
	"time"
)

type IPFSHandler struct {
	JsonHandler *JsonRpcHandler
	ipfsService serviceInterface.IpfsService
}

func NewIPFSCluster(
	chain constant.Chain,
	app *app.App,
) *IPFSHandler {
	return &IPFSHandler{
		JsonHandler: NewJsonRpcHandler(chain, nil, nil, nil, nil, app),
		ipfsService: app.IPFSSrv,
	}
}

func (h *IPFSHandler) newLogger(c echo.Context) *zap.Logger {
	// add chain name
	return h.JsonHandler.logger.With(zap.String("chain", h.JsonHandler.chain.Name), zap.String("request_id", c.Request().Context().Value("request_id").(string)))
}

func (h *IPFSHandler) Upload(c echo.Context) error {
	logger := h.newLogger(c)

	apiKey, err := h.JsonHandler.bindApiKey(c)
	if err != nil {
		return err
	}

	if rlErr := h.JsonHandler.rateLimit(c.Request().Context(), logger, apiKey, 1); rlErr != nil {
		logger.Debug("rate limit", zap.String("apiKey", apiKey), zap.Error(rlErr))
		return c.JSON(200, rlErr)
	}

	// Multipart form
	form, err := c.MultipartForm()
	if err != nil {
		return c.JSON(400, err)
	}
	requestFiles := form.File["files"]
	if len(requestFiles) == 0 {
		return c.JSON(200, nil)
	} else if len(requestFiles) == 1 {
		cid, err := h.ipfsService.Upload(c.Request().Context(), apiKey, requestFiles[0])
		if err != nil {
			return c.JSON(400, err)
		}
		return c.JSON(200, cid)
	}

	fs := make(map[string]files.Node)

	for _, f := range requestFiles {
		fmt.Println(f.Size)
		//sf := NewMapDirectory(map[string]Node{
		//	"file.txt": NewBytesFile([]byte(text)),
		//	"boop": NewMapDirectory(map[string]Node{
		//		"a.txt": NewBytesFile([]byte("bleep")),
		//		"b.txt": NewBytesFile([]byte("bloop")),
		//	}),
		//	"beep.txt": NewBytesFile([]byte("beep")),
		//})
	}

	sf := files.NewMapDirectory(fs)

	cid, err := h.ipfsService.UploadDir(c.Request().Context(), apiKey, files.NewMultiFileReader(sf, true))
	if err != nil {
		return c.JSON(400, err)
	}

	return c.JSON(200, cid)
}

func (h *IPFSHandler) List(c echo.Context) error {
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

	var fileList []models.IPFSFile
	err = h.ipfsService.ListUserFile(ctx, apiKey, &fileList)
	if err != nil {
		return c.JSON(400, err)
	}
	return c.JSON(200, fileList)
}

func (h *IPFSHandler) Get(c echo.Context) error {
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

func (h *IPFSHandler) Pin(c echo.Context) error {
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
