// Copyright (c) 2020, The OTNS Authors.
// All rights reserved.
//
// Redistribution and use in source and binary forms, with or without
// modification, are permitted provided that the following conditions are met:
// 1. Redistributions of source code must retain the above copyright
//    notice, this list of conditions and the following disclaimer.
// 2. Redistributions in binary form must reproduce the above copyright
//    notice, this list of conditions and the following disclaimer in the
//    documentation and/or other materials provided with the distribution.
// 3. Neither the name of the copyright holder nor the
//    names of its contributors may be used to endorse or promote products
//    derived from this software without specific prior written permission.
//
// THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS "AS IS"
// AND ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT LIMITED TO, THE
// IMPLIED WARRANTIES OF MERCHANTABILITY AND FITNESS FOR A PARTICULAR PURPOSE
// ARE DISCLAIMED. IN NO EVENT SHALL THE COPYRIGHT HOLDER OR CONTRIBUTORS BE
// LIABLE FOR ANY DIRECT, INDIRECT, INCIDENTAL, SPECIAL, EXEMPLARY, OR
// CONSEQUENTIAL DAMAGES (INCLUDING, BUT NOT LIMITED TO, PROCUREMENT OF
// SUBSTITUTE GOODS OR SERVICES; LOSS OF USE, DATA, OR PROFITS; OR BUSINESS
// INTERRUPTION) HOWEVER CAUSED AND ON ANY THEORY OF LIABILITY, WHETHER IN
// CONTRACT, STRICT LIABILITY, OR TORT (INCLUDING NEGLIGENCE OR OTHERWISE)
// ARISING IN ANY WAY OUT OF THE USE OF THIS SOFTWARE, EVEN IF ADVISED OF THE
// POSSIBILITY OF SUCH DAMAGE.

package web

import (
	"fmt"
	"os/exec"
	"runtime"

	"github.com/openthread/ot-ns/progctx"

	"github.com/openthread/ot-ns/logger"
)

var (
	grpcWebProxyProc *exec.Cmd

	grpcWebProxyParams struct {
		serverBindAddress   string
		serverHttpDebugPort int
		grpcServicePort     int
		webSitePort         int
	}
)

func ConfigWeb(serverBindAddress string, serverHttpDebugPort int, grpcServicePort int, webSitePort int) {
	grpcWebProxyParams.serverBindAddress = serverBindAddress
	grpcWebProxyParams.serverHttpDebugPort = serverHttpDebugPort
	grpcWebProxyParams.grpcServicePort = grpcServicePort
	grpcWebProxyParams.webSitePort = webSitePort
	logger.Debugf("ConfigWeb: %+v", grpcWebProxyParams)
}

func OpenWeb(ctx *progctx.ProgCtx) error {
	if err := assureGrpcWebProxyRunning(ctx); err != nil {
		logger.Errorf("start grpcwebproxy failed: %v", err)
		logger.Errorf("Web visualization is unusable. Please make sure grpcwebproxy is installed.")
		return err
	}

	return openWebBrowser(fmt.Sprintf("http://localhost:%d/visualize?addr=localhost:%d", grpcWebProxyParams.webSitePort, grpcWebProxyParams.serverHttpDebugPort))
}

// open opens the specified URL in the default browser of the user.
func openWebBrowser(url string) error {
	var cmd string
	var args []string

	switch runtime.GOOS {
	case "windows":
		cmd = "cmd"
		args = []string{"/c", "start"}
	case "darwin":
		cmd = "open"
	default: // "linux", "freebsd", "openbsd", "netbsd"
		cmd = "xdg-open"
	}

	args = append(args, url)
	return exec.Command(cmd, args...).Start()
}

func startGrpcWebProxy(ctx *progctx.ProgCtx) error {
	grpcWebProxyProc = exec.CommandContext(ctx, "grpcwebproxy", []string{
		fmt.Sprintf("--backend_addr=localhost:%d", grpcWebProxyParams.grpcServicePort),
		"--run_tls_server=false",
		"--allow_all_origins",
		"--server_http_max_read_timeout=1h",
		"--server_http_max_write_timeout=1h",
		fmt.Sprintf("--server_bind_address=%s", grpcWebProxyParams.serverBindAddress),
		fmt.Sprintf("--server_http_debug_port=%d", grpcWebProxyParams.serverHttpDebugPort),
	}...)

	return grpcWebProxyProc.Start()
}

func assureGrpcWebProxyRunning(ctx *progctx.ProgCtx) error {
	if grpcWebProxyProc == nil {
		if err := startGrpcWebProxy(ctx); err != nil {
			return err
		}

		logger.Infof("grpcwebproxy started.")

		ctx.WaitAdd("grpcwebproxy", 1)
		go func() {
			defer ctx.WaitDone("grpcwebproxy")

			err := grpcWebProxyProc.Wait()
			if err != nil && ctx.Err() == nil {
				logger.Warnf("grpcwebproxy exit unexpectedly: %v, try restarting ...", err)

				tryKillExistingGrpcWebProxyProcess()

				if err = startGrpcWebProxy(ctx); err != nil {
					logger.Errorf("grpcwebproxy restart failed: %v", err)
					return
				}

				logger.Infof("grpcwebproxy restarted.")
				err = grpcWebProxyProc.Wait()

				if err != nil && ctx.Err() == nil {
					logger.Errorf("grpcwebproxy exit unexpectedly: %v", err)
					logger.Errorf("Web visualization might not be working properly!")
				}
			}
		}()
	}

	return nil
}

func tryKillExistingGrpcWebProxyProcess() {
	var err error

	defer func() {
		if err != nil {
			logger.Warnf("Kill existing grpcwebproxy process failed: %v", err)
		}
	}()

	pattern := fmt.Sprintf("grpcwebproxy.*--server_http_debug_port=%d", grpcWebProxyParams.serverHttpDebugPort)
	cmd := exec.Command("pkill", "-f", pattern)
	if err = cmd.Start(); err != nil {
		logger.Errorf("pkill grpcwebproxy failed: %v", err)
		return
	}

	if err = cmd.Wait(); err != nil || !cmd.ProcessState.Success() {
		logger.Errorf("pkill grpcwebproxy failed: %s", cmd.ProcessState)
	}
}
