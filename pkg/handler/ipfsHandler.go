package handler

import (
	"context"
	"fmt"
	files "github.com/ipfs/go-ipfs-files"
	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
	"mime/multipart"
	"net/http"
	"starnet/chain-api/pkg/app"
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

// Upload TODO: split to a UploadFile & UploadDir
func (h *IPFSHandler) Upload(c echo.Context) error {
	logger := h.newLogger(c)

	apiKey, err := h.JsonHandler.bindApiKey(c)
	if err != nil {
		return err
	}

	if rlErr := h.JsonHandler.rateLimit(c.Request().Context(), logger, apiKey, 1); rlErr != nil {
		logger.Debug("rate limit", zap.String("apiKey", apiKey), zap.Error(rlErr))
		return c.JSON(http.StatusOK, rlErr)
	}

	// Multipart form
	form, err := c.MultipartForm()
	if err != nil {
		return c.JSON(http.StatusBadRequest, err)
	}
	requestFiles := form.File["files"]

	if len(requestFiles) == 0 {
		return c.JSON(http.StatusOK, nil)
	} else if len(requestFiles) == 1 {
		cid, err := h.ipfsService.Upload(c.Request().Context(), apiKey, requestFiles[0])
		if err != nil {
			return c.JSON(http.StatusBadRequest, err)
		}
		return c.JSON(http.StatusOK, cid)
	}

	var rootDir string
	rootDir = "root"
	//for _, f := range requestFiles {
	//	if strings.Contains(f.Filename, "/") {
	//		rootDir = strings.Split(f.Filename, "/")[0]
	//		break
	//	}
	//}
	fs := make(map[string]files.Node)
	fileMap := make(map[string]*multipart.FileHeader)
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

	//sf := files.NewMapDirectory(map[string]files.Node{
	//	rootDir: files.NewMapDirectory(fs),
	//})
	sf := files.NewMapDirectory(fs)

	cid, err := h.ipfsService.UploadDir(c.Request().Context(), apiKey, files.NewMultiFileReader(sf, true))
	if err != nil {
		return c.JSON(http.StatusBadRequest, err)
	}

	return c.JSON(http.StatusOK, cid)
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
		return c.JSON(http.StatusOK, rlErr)
	}

	ctx, cancelFunc := context.WithTimeout(c.Request().Context(), time.Second*5)
	defer cancelFunc()

	// TODO: Get
	// proxy request to ipfs gateway
	fmt.Println(ctx.Err())

	// TODO: add file size compute
	return c.JSONBlob(http.StatusOK, nil)
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

	// TODO: Pin
	fmt.Println(ctx.Err())

	// FIXME: resp cid
	return c.JSON(http.StatusOK, nil)
}
