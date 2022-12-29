package handler

import (
	"context"
	"fmt"
	files "github.com/ipfs/go-ipfs-files"
	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
	"io"
	"mime/multipart"
	"net/http"
	"starnet/chain-api/pkg/app"
	"starnet/chain-api/pkg/request"
	ratelimitv1 "starnet/chain-api/ratelimit/v1"
	serviceInterface "starnet/chain-api/service/interface"
	"starnet/starnet/constant"
	"starnet/starnet/models"
	"strings"
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

// Add .
func (h *IPFSHandler) Add(c echo.Context) error {
	logger := h.newLogger(c)

	apiKey, err := h.JsonHandler.bindApiKey(c)
	if err != nil {
		return err
	}

	if rlErr := h.JsonHandler.rateLimit(c.Request().Context(), logger, apiKey, 1); rlErr != nil {
		logger.Debug("rate limit", zap.String("apiKey", apiKey), zap.Error(rlErr))
		return c.JSON(http.StatusOK, rlErr)
	}

	var params = request.AddParam{}
	if err := c.Bind(&params); err != nil {
		return c.JSON(http.StatusBadRequest, err)
	}

	// Multipart form
	form, err := c.MultipartForm()
	if err != nil {
		logger.Error("param bind error", zap.Error(err))
		return c.JSON(http.StatusBadRequest, err)
	}
	requestFiles := form.File["file"]

	var rootDir string
	var dataSize int64
	for _, f := range requestFiles {
		if strings.Contains(f.Filename, "/") {
			rootDir = "root"
		}
		dataSize += f.Size
	}

	// bandwidth use check
	if err := h.JsonHandler.rateLimiter.BandwidthHook(c.Request().Context(), h.JsonHandler.chain.ChainID, apiKey, dataSize, ratelimitv1.BandWidthUpload); err != nil {
		return c.JSON(http.StatusBadRequest, err)
	}

	fs := make(map[string]files.Node)
	fileMap := make(map[string]*multipart.FileHeader)
	if rootDir != "" && params.WrapWithDirectory {
		for _, f := range requestFiles {
			fileMap[rootDir+"-"+f.Filename] = f
		}
		tree := processFileMap(rootDir, fileMap)
		for _, node := range tree.Files {
			fmt.Println(node)
			if node.FileData == nil {
				fs[node.Name], err = newMapDirectory(node.Files)
				if err != nil {
					return c.JSON(http.StatusBadRequest, err)
				}
			} else {
				file, err := node.FileData.Open()
				if err != nil {
					return c.JSON(http.StatusBadRequest, err)
				}
				fs[node.Name] = files.NewReaderFile(file)
			}
		}
	} else {
		for _, f := range requestFiles {
			file, err := f.Open()
			if err != nil {
				return c.JSON(http.StatusBadRequest, err)
			}
			fs := make(map[string]files.Node)
			fs[f.Filename] = files.NewReaderFile(file)
		}
	}

	sf := files.NewMapDirectory(fs)

	results, err := h.ipfsService.Add(c.Request().Context(), apiKey, params, files.NewMultiFileReader(sf, true))
	if err != nil {
		return c.JSON(http.StatusBadRequest, err)
	}

	return c.JSON(http.StatusOK, results)
}

func newMapDirectory(teeFiles map[string]fileTree) (files.Directory, error) {
	var err error
	filesNode := make(map[string]files.Node)
	for _, node := range teeFiles {
		fmt.Println(node)
		if node.FileData == nil {
			filesNode[node.Name], err = newMapDirectory(node.Files)
			if err != nil {
				return nil, err
			}
		} else {
			file, err := node.FileData.Open()
			if err != nil {
				return nil, err
			}
			filesNode[node.Name] = files.NewReaderFile(file)
		}
	}
	return files.NewMapDirectory(filesNode), nil
}

type fileTree struct {
	Name     string
	FileData *multipart.FileHeader
	Files    map[string]fileTree
}

func processFileMap(rootDir string, m map[string]*multipart.FileHeader) fileTree {
	root := fileTree{
		Name:     rootDir,
		FileData: nil,
		Files:    make(map[string]fileTree),
	}

	for path, data := range m {
		pathSegments := strings.Split(path, "/")
		processPathSegments(root, pathSegments, data)
	}

	return root
}

func processPathSegments(root fileTree, pathSegments []string, data *multipart.FileHeader) {
	currentSegment := pathSegments[0]

	if len(pathSegments) == 1 {
		root.Files[currentSegment] = fileTree{
			Name:     currentSegment,
			FileData: data,
			Files:    nil,
		}
		return
	}

	subtree, ok := root.Files[currentSegment]
	if !ok {
		subtree = fileTree{
			Name:     currentSegment,
			FileData: nil,
			Files:    make(map[string]fileTree),
		}
		root.Files[currentSegment] = subtree
	}
	processPathSegments(subtree, pathSegments[1:], data)
}

func (h *IPFSHandler) List(c echo.Context) error {
	logger := h.newLogger(c)

	apiKey, err := h.JsonHandler.bindApiKey(c)
	if err != nil {
		return err
	}

	if rlErr := h.JsonHandler.rateLimit(c.Request().Context(), logger, apiKey, 1); rlErr != nil {
		logger.Debug("rate limit", zap.String("apiKey", apiKey), zap.Error(rlErr))
		return c.JSON(http.StatusOK, rlErr)
	}

	ctx, cancelFunc := context.WithTimeout(c.Request().Context(), time.Second*5)
	defer cancelFunc()

	var fileList []models.IPFSFile
	err = h.ipfsService.ListUserFile(ctx, apiKey, &fileList)
	if err != nil {
		return c.JSON(http.StatusBadRequest, err)
	}
	return c.JSON(http.StatusOK, fileList)
}

func (h *IPFSHandler) Get(c echo.Context) error {
	logger := h.newLogger(c)

	apiKey, err := h.JsonHandler.bindApiKey(c)
	if err != nil {
		return err
	}

	if rlErr := h.JsonHandler.rateLimit(c.Request().Context(), logger, apiKey, 1); rlErr != nil {
		logger.Debug("rate limit", zap.String("apiKey", apiKey), zap.Error(rlErr))
		return c.JSON(http.StatusBadRequest, rlErr)
	}

	var cid string
	err = echo.PathParamsBinder(c).String("cid", &cid).BindError()
	if err != nil {
		return c.JSON(http.StatusBadRequest, err)
	}

	object, err := h.ipfsService.GetObject(cid)
	if err != nil {
		return c.JSON(http.StatusBadRequest, err)
	}
	data, err := io.ReadAll(object)
	if err != nil {
		return c.JSON(http.StatusBadRequest, err)
	}

	// bandwidth use check
	if err := h.JsonHandler.rateLimiter.BandwidthHook(c.Request().Context(), h.JsonHandler.chain.ChainID, apiKey, int64(len(data)), ratelimitv1.BandWidthDownload); err != nil {
		return c.JSON(http.StatusBadRequest, err)
	}
	return c.JSONBlob(http.StatusOK, data)
}

func (h *IPFSHandler) Pin(c echo.Context) error {
	logger := h.newLogger(c)

	apiKey, err := h.JsonHandler.bindApiKey(c)
	if err != nil {
		return err
	}

	if rlErr := h.JsonHandler.rateLimit(c.Request().Context(), logger, apiKey, 1); rlErr != nil {
		logger.Debug("rate limit", zap.String("apiKey", apiKey), zap.Error(rlErr))
		return c.JSON(http.StatusOK, rlErr)
	}

	ctx, cancelFunc := context.WithTimeout(c.Request().Context(), time.Second*5)
	defer cancelFunc()

	var cid string
	err = echo.PathParamsBinder(c).String("cid", &cid).BindError()
	if err != nil {
		return c.JSON(http.StatusBadRequest, err)
	}

	err = h.ipfsService.Pin(ctx, cid)
	if err != nil {
		return c.JSON(http.StatusBadRequest, err)
	}

	return c.JSON(http.StatusOK, nil)
}

func (h *IPFSHandler) Proxy(c echo.Context) error {
	logger := h.newLogger(c)

	apiKey, err := h.JsonHandler.bindApiKey(c)
	if err != nil {
		return c.JSON(http.StatusBadRequest, err)
	}

	if rlErr := h.JsonHandler.rateLimit(c.Request().Context(), logger, apiKey, 1); rlErr != nil {
		logger.Debug("rate limit", zap.String("apiKey", apiKey), zap.Error(rlErr))
		return c.JSON(http.StatusBadRequest, rlErr)
	}

	w := c.Response().Writer
	r := c.Request()

	// Create a new HTTP client
	client := &http.Client{}

	// Create a new request to the target URL FIXME: request base url
	targetReq, err := http.NewRequest(r.Method, "http://127.0.0.1:9095"+r.URL.Path, r.Body)
	if err != nil {
		return c.JSON(http.StatusBadRequest, err)
	}

	// Copy the request headers to the new request
	targetReq.Header = r.Header

	// Forward the request to the target URL
	targetResp, err := client.Do(targetReq)
	if err != nil {
		return c.JSON(http.StatusBadRequest, err)
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			logger.Error("close the targetResp err:", zap.Error(err))
		}
	}(targetResp.Body)

	// Copy the response headers to the original response
	for k, v := range targetResp.Header {
		w.Header()[k] = v
	}
	w.WriteHeader(targetResp.StatusCode)

	// Copy the response body to the original response
	// TODO: read the data from the body
	body, err := io.ReadAll(targetResp.Body)
	if err != nil {
		return c.JSON(http.StatusBadRequest, err)
	}
	_, err = w.Write(body)
	if err != nil {
		return c.JSON(http.StatusBadRequest, err)
	}

	// bandwidth use check
	// TODO: r.URL.Path -> BandWidthType
	//if err := h.JsonHandler.rateLimiter.BandwidthHook(c.Request().Context(), h.JsonHandler.chain.ChainID, apiKey, int64(len(body))); err != nil {
	//	return c.JSON(http.StatusBadRequest, err)
	//}
	return nil
}
