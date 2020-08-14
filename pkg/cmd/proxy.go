/*
Copyright 2020 The KubeOne Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package cmd

import (
	"fmt"
	"io"
	"net/http"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	"k8c.io/kubeone/pkg/state"
)

type proxyOpts struct {
	globalOptions
	ListenAddr string `longflag:"listen"`
}

func proxyCmd(rootFlags *pflag.FlagSet) *cobra.Command {
	opts := &proxyOpts{}

	cmd := &cobra.Command{
		Use:   "proxy",
		Short: "Proxy to the kube-apiserver using SSH tunnel",
		Long: `
HTTPS Proxy (CONNECT method) SSH tunnel.

This command helps to reach kubeapi endpoint with local kubectl in case when private/firewalled endpoint is used (e.g.
internal loadbalancer). It creates SSH tunnel to one of the control-plane nodes and then proxies incomming requests
through it.
`,
		Example: `kubeone proxy -m mycluster.yaml -t terraformoutput.json`,
		RunE: func(*cobra.Command, []string) error {
			gopts, err := persistentGlobalOptions(rootFlags)
			if err != nil {
				return errors.Wrap(err, "unable to get global flags")
			}
			opts.globalOptions = *gopts

			return setupProxyTunnel(opts)
		},
	}

	cmd.Flags().StringVar(&opts.ListenAddr, longFlagName(opts, "ListenAddr"), "127.0.0.1:8888", "SSH tunnel HTTP proxy bind address")

	return cmd
}

func setupProxyTunnel(opts *proxyOpts) error {
	s, err := opts.BuildState()
	if err != nil {
		return err
	}

	server := &http.Server{
		Addr: opts.ListenAddr,
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method != http.MethodConnect {
				http.Error(w, "bad request", http.StatusBadRequest)
				return
			}

			if err := handleTunneling(w, r, s); err != nil {
				code := http.StatusInternalServerError
				if err1, ok := err.(*httpError); ok {
					code = err1.code
				}
				http.Error(w, err.Error(), code)
			}
		}),
	}

	fmt.Println("SSH tunnel started, please open another terminal and setup environment")
	fmt.Printf("export HTTPS_PROXY=http://%s\n", opts.ListenAddr)
	return server.ListenAndServe()
}

type httpError struct {
	err  error
	code int
}

func (e *httpError) Error() string {
	return fmt.Sprintf("error: %s, code: %d", e.err, e.code)
}

func handleTunneling(w http.ResponseWriter, r *http.Request, s *state.State) error {
	tunn, err := s.Connector.Tunnel(s.Cluster.RandomHost())
	if err != nil {
		return &httpError{err: err, code: http.StatusServiceUnavailable}
	}

	destConn, err := tunn.TunnelTo(s.Context, "tcp4", r.Host)
	if err != nil {
		tunn.Close()
		return &httpError{err: err, code: http.StatusServiceUnavailable}
	}

	hijacker, ok := w.(http.Hijacker)
	if !ok {
		return &httpError{err: err, code: http.StatusInternalServerError}
	}

	w.WriteHeader(http.StatusOK)
	clientConn, _, err := hijacker.Hijack()
	if err != nil {
		return &httpError{err: err, code: http.StatusServiceUnavailable}
	}

	go func() {
		if err := iocopy(destConn, clientConn); err != nil {
			s.Logger.Errorf("%v", err)
		}
	}()

	go func() {
		if err := iocopy(clientConn, destConn); err != nil {
			s.Logger.Errorf("%v", err)
		}
	}()

	return nil
}

func iocopy(dst io.WriteCloser, src io.ReadCloser) error {
	defer dst.Close()
	defer src.Close()

	_, err := io.Copy(dst, src)
	return err
}
