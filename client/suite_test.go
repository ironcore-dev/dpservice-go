/*
Copyright 2022 The Metal Authors.

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

package client

import (
	"context"
	"testing"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	dpdkproto "github.com/onmetal/net-dpservice-go/proto"
	//+kubebuilder:scaffold:imports
)

// These tests use Ginkgo (BDD-style Go testing framework). Refer to
// http://onsi.github.io/ginkgo/ to learn more about Ginkgo.

var (
	ctxCancel       context.CancelFunc
	ctxGrpc         context.Context
	dpserviceAddr   string = "127.0.0.1:1337"
	dpdkProtoClient dpdkproto.DPDKonmetalClient
	dpdkClient      Client
)

// This assumes running dp-service on the same localhost of this test suite
// /test/dp_service.py --no-init

func TestGrpcFuncs(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Controller Suite")
}

var _ = BeforeSuite(func() {

	//+kubebuilder:scaffold:scheme

	// setup net-dpservice client
	ctxGrpc, ctxCancel = context.WithTimeout(context.Background(), 100*time.Millisecond)

	conn, err := grpc.DialContext(ctxGrpc, dpserviceAddr, grpc.WithTransportCredentials(insecure.NewCredentials()), grpc.WithBlock())
	Expect(err).NotTo(HaveOccurred())

	dpdkProtoClient = dpdkproto.NewDPDKonmetalClient(conn)
	dpdkClient = NewClient(dpdkProtoClient)

	_, err = dpdkClient.Initialize(context.TODO())
	Expect(err).NotTo(HaveOccurred())
})

var _ = AfterSuite(func() {

})
