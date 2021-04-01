// Copyright 2020 The Swarm Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package api_test

import (
	"bytes"
	// "io/ioutil"
	"net/http"
	"os"
	"sort"
	"testing"

	"github.com/ethersphere/bee/pkg/api"
	"github.com/ethersphere/bee/pkg/jsonhttp"
	"github.com/ethersphere/bee/pkg/jsonhttp/jsonhttptest"
	"github.com/ethersphere/bee/pkg/logging"
	statestore "github.com/ethersphere/bee/pkg/statestore/mock"
	"github.com/ethersphere/bee/pkg/storage/mock"
	"github.com/ethersphere/bee/pkg/swarm"
	"github.com/ethersphere/bee/pkg/tags"
	"github.com/ethersphere/bee/pkg/traversal"
)

func TestPinFilesHandler(t *testing.T) {
	var (
		fileUploadResource      = "/bzz"
		pinFilesResource        = "/pin/files"
		pinFilesAddressResource = func(addr string) string { return pinFilesResource + "/" + addr }
		pinChunksResource       = "/pin/chunks"

		simpleData = []byte("this is a simple text")

		mockStorer       = mock.NewStorer()
		mockStatestore   = statestore.NewStateStore()
		traversalService = traversal.NewService(mockStorer)
		logger           = logging.New(os.Stdout, 6)
		client, _, _     = newTestServer(t, testServerOptions{
			Storer:    mockStorer,
			Traversal: traversalService,
			Tags:      tags.NewTags(mockStatestore, logger),
			Logger:    logger,
		})
	)

	t.Run("pin-file-1", func(t *testing.T) {
		rootHash := "dd13a5a6cc9db3ef514d645e6719178dbfb1a90b49b9262cafce35b0d27cf245"
		metadataHash := "d936d7180f230b3424842ea10848aa205f2f0e830cb9cc7588a39c9381544bf9"
		contentHash := "838d0a193ecd1152d1bb1432d5ecc02398533b2494889e23b8bd5ace30ac2aeb"

		jsonhttptest.Request(t, client, http.MethodPost,
			fileUploadResource+"?name=somefile.txt", http.StatusOK,
			jsonhttptest.WithRequestBody(bytes.NewReader(simpleData)),
			jsonhttptest.WithExpectedJSONResponse(api.FileUploadResponse{
				Reference: swarm.MustParseHexAddress(rootHash),
			}),
			jsonhttptest.WithRequestHeader("Content-Type", "text/plain"),
		)

		jsonhttptest.Request(t, client, http.MethodPost, pinFilesAddressResource(rootHash), http.StatusOK,
			jsonhttptest.WithExpectedJSONResponse(jsonhttp.StatusResponse{
				Message: http.StatusText(http.StatusOK),
				Code:    http.StatusOK,
			}),
		)

		hashes := []string{rootHash, metadataHash, contentHash}
		sort.Strings(hashes)

		expectedResponse := api.ListPinnedChunksResponse{
			Chunks: []api.PinnedChunk{},
		}

		for _, h := range hashes {
			expectedResponse.Chunks = append(expectedResponse.Chunks, api.PinnedChunk{
				Address:    swarm.MustParseHexAddress(h),
				PinCounter: 1,
			})
		}

		jsonhttptest.Request(t, client, http.MethodGet, pinChunksResource, http.StatusOK,
			jsonhttptest.WithExpectedJSONResponse(expectedResponse),
		)
	})

	t.Run("unpin-file-1", func(t *testing.T) {
		rootHash := "dd13a5a6cc9db3ef514d645e6719178dbfb1a90b49b9262cafce35b0d27cf245"

		jsonhttptest.Request(t, client, http.MethodDelete, pinFilesAddressResource(rootHash), http.StatusOK,
			jsonhttptest.WithExpectedJSONResponse(jsonhttp.StatusResponse{
				Message: http.StatusText(http.StatusOK),
				Code:    http.StatusOK,
			}),
		)

		jsonhttptest.Request(t, client, http.MethodGet, pinChunksResource, http.StatusOK,
			jsonhttptest.WithExpectedJSONResponse(api.ListPinnedChunksResponse{
				Chunks: []api.PinnedChunk{},
			}),
		)
	})

}
