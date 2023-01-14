package service

import (
	"context"
	"io"

	"starnet/chain-api/pkg/request"
	"starnet/chain-api/pkg/response"
	serviceInterface "starnet/chain-api/service/interface"
	"starnet/starnet/models"

	"github.com/ipfs-cluster/ipfs-cluster/api"
	files "github.com/ipfs/go-ipfs-files"
	"github.com/spf13/cast"
)

var _ serviceInterface.IpfsService = &IpfsService{}

func (s *IpfsService) CheckMethod(pathStr string) bool {
	for _, v := range s.allowMethod {
		if v == pathStr {
			return true
		}
	}
	return false
}

// AddCluster file or directory to ipfs cluster
func (s *IpfsService) AddCluster(ctx context.Context, apiKey string, apiParam request.AddParam, multiFileR *files.MultiFileReader) ([]response.AddResp, error) {
	out := make(chan api.AddedOutput)
	e := make(chan error)
	userID := s.getUserIDByAPIKey(ctx, apiKey)
	go func() {
		param := api.AddParams{
			Recursive: apiParam.Recursive,
			Hidden:    apiParam.Hidden,
			Wrap:      apiParam.WrapWithDirectory,
			NoPin:     !apiParam.Pin,
			IPFSAddParams: api.IPFSAddParams{
				Chunker:    apiParam.Chunker,
				Progress:   apiParam.Progress,
				CidVersion: apiParam.CidVersion,
				NoCopy:     apiParam.Nocopy,
			},
		}
		e <- s.client.AddMultiFile(ctx, multiFileR, param, out)
	}()
	var dirHash string
	var results []response.AddResp
	var dbFiles []*models.IPFSFile
	pinStatus := models.PinStatusUnPin
	if apiParam.Pin {
		pinStatus = models.PinStatusPin
	}
	for {
		select {
		case err := <-e:
			if err != nil {
				return nil, err
			}
			goto SAVED
		case result := <-out:
			if result.Cid.String() != "b" {
				hash := result.Cid.String()
				if result.Name == "" {
					dirHash = hash
				}
				results = append(results, response.AddResp{
					Name: result.Name,
					Hash: hash,
					Size: cast.ToString(result.Size),
				})
				dbFiles = append(dbFiles, &models.IPFSFile{
					PinStatus: pinStatus,
					UserId:    userID,
					FileSize:  result.Size,
					FileName:  result.Name,
					CID:       hash,
				})
			}
		}
	}
SAVED:
	if dirHash != "" {
		for k := range dbFiles {
			dbFiles[k].WrapWithDirCid = dirHash
			dbFiles[k].WrapDirName = apiParam.WrapDirName
		}
	}
	if err := s.ipfsDao.BatchSaveFiles(dbFiles); err != nil {
		return nil, err
	}
	return results, nil
}

func (s *IpfsService) GetObject(cidStr string) (io.ReadCloser, error) {
	return s.ipfsShell.Cat(cidStr)
}
