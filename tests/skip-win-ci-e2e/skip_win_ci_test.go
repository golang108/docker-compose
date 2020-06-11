/*
	Copyright (c) 2020 Docker Inc.

	Permission is hereby granted, free of charge, to any person
	obtaining a copy of this software and associated documentation
	files (the "Software"), to deal in the Software without
	restriction, including without limitation the rights to use, copy,
	modify, merge, publish, distribute, sublicense, and/or sell copies
	of the Software, and to permit persons to whom the Software is
	furnished to do so, subject to the following conditions:

	The above copyright notice and this permission notice shall be
	included in all copies or substantial portions of the Software.

	THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND,
	EXPRESS OR IMPLIED,
	INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
	FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT.
	IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT
	HOLDERS BE LIABLE FOR ANY CLAIM,
	DAMAGES OR OTHER LIABILITY,
	WHETHER IN AN ACTION OF CONTRACT,
	TORT OR OTHERWISE,
	ARISING FROM, OUT OF OR IN CONNECTION WITH
	THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.
*/

package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	. "github.com/onsi/gomega"
	"github.com/stretchr/testify/suite"

	. "github.com/docker/api/tests/framework"
)

type NonWinCIE2eSuite struct {
	Suite
}

func (s *NonWinCIE2eSuite) TestKillChildOnCancel() {
	It("should kill docker-classic if parent command is cancelled", func() {
		out := s.ListProcessesCommand().ExecOrDie()
		Expect(out).NotTo(ContainSubstring("docker-classic"))

		dir := s.ConfigDir
		Expect(ioutil.WriteFile(filepath.Join(dir, "Dockerfile"), []byte(`FROM alpine:3.10
RUN sleep 100`), 0644)).To(Succeed())
		shutdown := make(chan time.Time)
		errs := make(chan error)
		ctx := s.NewDockerCommand("build", "--no-cache", "-t", "test-sleep-image", ".").WithinDirectory(dir).WithTimeout(shutdown)
		go func() {
			_, err := ctx.Exec()
			errs <- err
		}()
		err := WaitFor(time.Second, 10*time.Second, errs, func() bool {
			out := s.ListProcessesCommand().ExecOrDie()
			return strings.Contains(out, "docker-classic")
		})
		Expect(err).NotTo(HaveOccurred())
		log.Println("Killing docker process")

		close(shutdown)
		err = WaitFor(time.Second, 12*time.Second, nil, func() bool {
			out := s.ListProcessesCommand().ExecOrDie()
			return !strings.Contains(out, "docker-classic")
		})
		Expect(err).NotTo(HaveOccurred())
	})
}

func (s *NonWinCIE2eSuite) TestAPIServer() {
	_, err := exec.LookPath("yarn")
	if err != nil || os.Getenv("SKIP_NODE") != "" {
		s.T().Skip("skipping, yarn not installed")
	}
	It("can run 'serve' command", func() {
		cName := "test-example"
		s.NewDockerCommand("context", "create", cName, "example").ExecOrDie()

		//sPath := fmt.Sprintf("unix:///%s/docker.sock", s.ConfigDir)
		sPath, cliAddress := s.getGrpcServerAndCLientAddress()
		server, err := serveAPI(s.ConfigDir, sPath)
		Expect(err).To(BeNil())
		defer killProcess(server)

		s.NewCommand("yarn", "install").WithinDirectory("../node-client").ExecOrDie()
		output := s.NewCommand("yarn", "run", "start", cName, cliAddress).WithinDirectory("../node-client").ExecOrDie()
		Expect(output).To(ContainSubstring("nginx"))
	})
}

func (s *NonWinCIE2eSuite) getGrpcServerAndCLientAddress() (string, string) {
	if IsWindows() {
		return "npipe:////./pipe/clibackend", "unix:////./pipe/clibackend"
	}
	socketName := fmt.Sprintf("unix:///%s/docker.sock", s.ConfigDir)
	return socketName, socketName
}

func TestE2e(t *testing.T) {
	suite.Run(t, new(NonWinCIE2eSuite))
}

func killProcess(process *os.Process) {
	err := process.Kill()
	Expect(err).To(BeNil())
}

func serveAPI(configDir string, address string) (*os.Process, error) {
	cmd := exec.Command("../../bin/docker", "--config", configDir, "serve", "--address", address)
	err := cmd.Start()
	if err != nil {
		return nil, err
	}
	return cmd.Process, nil
}
