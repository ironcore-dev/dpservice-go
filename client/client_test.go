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
	"net/netip"

	"github.com/onmetal/net-dpservice-go/api"
	"github.com/onmetal/net-dpservice-go/errors"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("interface", Label("interface"), func() {
	ctx := context.TODO()

	Context("When using interface functions", Ordered, func() {
		var res *api.Interface
		var err error

		It("should create successfully", func() {
			ipv4 := netip.MustParseAddr("10.200.1.4")
			ipv6 := netip.MustParseAddr("2000:200:1::4")
			iface := api.Interface{
				InterfaceMeta: api.InterfaceMeta{
					ID: "vm4",
				},
				Spec: api.InterfaceSpec{
					IPv4:   &ipv4,
					IPv6:   &ipv6,
					VNI:    200,
					Device: "net_tap5",
				},
			}
			res, err = dpdkClient.CreateInterface(ctx, &iface)
			Expect(err).ToNot(HaveOccurred())

			Expect(res.ID).To(Equal("vm4"))
			Expect(res.Spec.VNI).To(Equal(uint32(200)))
		})

		It("should not be created when already existing", func() {
			res, err = dpdkClient.CreateInterface(ctx, res)
			Expect(err).To(HaveOccurred())

			Expect(res.Status.Code).To(Equal(uint32(errors.ALREADY_EXISTS)))
		})

		It("should get successfully", func() {
			res, err = dpdkClient.GetInterface(ctx, res.ID)
			Expect(err).ToNot(HaveOccurred())

			Expect(res.Spec.IPv4.String()).To(Equal("10.200.1.4"))
			Expect(res.Spec.IPv6.String()).To(Equal("2000:200:1::4"))
		})

		It("should list successfully", func() {
			ifaces, err := dpdkClient.ListInterfaces(ctx)
			Expect(err).ToNot(HaveOccurred())

			Expect(len(ifaces.Items)).To(Equal(1))
			Expect(ifaces.Items[0].Kind).To(Equal("Interface"))
		})

		It("should delete successfully", func() {
			res, err = dpdkClient.DeleteInterface(ctx, res.ID)
			Expect(err).ToNot(HaveOccurred())

			res, err = dpdkClient.GetInterface(ctx, "vm4")
			Expect(err).To(HaveOccurred())
			Expect(res.Status.Code).To(Equal(uint32(errors.NOT_FOUND)))
		})
	})
})

var _ = Describe("interface related", Label("prefix", "lbprefix", "vip"), func() {
	ctx := context.TODO()

	// Creates the network interface object
	// OncePerOrdered decorator will run this only once per Ordered spec and not before every It spec
	BeforeEach(OncePerOrdered, func() {
		ipv4 := netip.MustParseAddr("10.200.1.4")
		ipv6 := netip.MustParseAddr("2000:200:1::4")
		iface := api.Interface{
			InterfaceMeta: api.InterfaceMeta{
				ID: "vm4",
			},
			Spec: api.InterfaceSpec{
				IPv4:   &ipv4,
				IPv6:   &ipv6,
				VNI:    200,
				Device: "net_tap5",
			},
		}
		_, err := dpdkClient.CreateInterface(ctx, &iface)
		Expect(err).ToNot(HaveOccurred())

		// Deletes the network interface object after spec is completed
		DeferCleanup(func(ctx SpecContext) {
			_, err := dpdkClient.DeleteInterface(ctx, "vm4")
			Expect(err).ToNot(HaveOccurred())
		})
	})

	Context("When using prefix functions", Ordered, func() {
		var res *api.Prefix
		var err error

		It("should create successfully", func() {
			prefix := api.Prefix{
				PrefixMeta: api.PrefixMeta{
					InterfaceID: "vm4",
				},
				Spec: api.PrefixSpec{
					Prefix: netip.MustParsePrefix("10.20.30.0/24"),
				},
			}

			res, err = dpdkClient.CreatePrefix(ctx, &prefix)
			Expect(err).ToNot(HaveOccurred())

			Expect(res.InterfaceID).To(Equal("vm4"))
			Expect(res.Spec.Prefix.String()).To(Equal("10.20.30.0/24"))
		})

		It("should not be created when already existing", func() {
			res, err = dpdkClient.CreatePrefix(ctx, res)
			Expect(err).To(HaveOccurred())

			Expect(res.Status.Code).To(Equal(uint32(errors.ROUTE_EXISTS)))
		})

		It("should list successfully", func() {
			prefixes, err := dpdkClient.ListPrefixes(ctx, "vm4")
			Expect(err).ToNot(HaveOccurred())

			Expect(len(prefixes.Items)).To(Equal(1))
			Expect(prefixes.Items[0].Kind).To(Equal("Prefix"))
		})

		It("should delete successfully", func() {
			res, err = dpdkClient.DeletePrefix(ctx, res.InterfaceID, &res.Spec.Prefix)
			Expect(err).ToNot(HaveOccurred())

			prefixes, err := dpdkClient.ListPrefixes(ctx, "vm4")
			Expect(err).ToNot(HaveOccurred())
			Expect(len(prefixes.Items)).To(Equal(0))
		})
	})

	Context("When using lbprefix functions", Ordered, func() {
		var res *api.LoadBalancerPrefix
		var err error

		It("should create successfully", func() {
			lbprefix := api.LoadBalancerPrefix{
				LoadBalancerPrefixMeta: api.LoadBalancerPrefixMeta{
					InterfaceID: "vm4",
				},
				Spec: api.LoadBalancerPrefixSpec{
					Prefix: netip.MustParsePrefix("10.10.10.0/24"),
				},
			}

			res, err = dpdkClient.CreateLoadBalancerPrefix(ctx, &lbprefix)
			Expect(err).ToNot(HaveOccurred())

			Expect(res.InterfaceID).To(Equal("vm4"))
			Expect(res.Spec.Prefix.String()).To(Equal("10.10.10.0/24"))
		})

		It("should not be created when already existing", func() {
			res, err = dpdkClient.CreateLoadBalancerPrefix(ctx, res)
			Expect(err).To(HaveOccurred())

			Expect(res.Status.Code).To(Equal(uint32(errors.ALREADY_EXISTS)))
		})

		It("should list successfully", func() {
			lbprefixes, err := dpdkClient.ListLoadBalancerPrefixes(ctx, "vm4")
			Expect(err).ToNot(HaveOccurred())

			Expect(len(lbprefixes.Items)).To(Equal(1))
			Expect(lbprefixes.Items[0].Kind).To(Equal("LoadBalancerPrefix"))
		})

		It("should delete successfully", func() {
			res, err = dpdkClient.DeleteLoadBalancerPrefix(ctx, res.InterfaceID, &res.Spec.Prefix)
			Expect(err).ToNot(HaveOccurred())

			lbprefixes, err := dpdkClient.ListLoadBalancerPrefixes(ctx, "vm4")
			Expect(err).ToNot(HaveOccurred())
			Expect(len(lbprefixes.Items)).To(Equal(0))
		})
	})

	Context("When using virtualIP functions", Ordered, func() {
		var res *api.VirtualIP
		var err error

		It("should create successfully", func() {
			ip := netip.MustParseAddr("20.20.20.20")
			vip := api.VirtualIP{
				VirtualIPMeta: api.VirtualIPMeta{
					InterfaceID: "vm4",
				},
				Spec: api.VirtualIPSpec{
					IP: &ip,
				},
			}

			res, err = dpdkClient.CreateVirtualIP(ctx, &vip)
			Expect(err).ToNot(HaveOccurred())

			Expect(res.InterfaceID).To(Equal("vm4"))
			Expect(res.Spec.IP.String()).To(Equal("20.20.20.20"))
		})

		It("should not be created when already existing", func() {
			res, err = dpdkClient.CreateVirtualIP(ctx, res)
			Expect(err).To(HaveOccurred())

			Expect(res.Status.Code).To(Equal(uint32(errors.SNAT_EXISTS)))
		})

		It("should get successfully", func() {
			res, err = dpdkClient.GetVirtualIP(ctx, "vm4")
			Expect(err).ToNot(HaveOccurred())

			Expect(res.InterfaceID).To(Equal("vm4"))
			Expect(res.Spec.UnderlayRoute).ToNot(BeNil())
		})

		It("should delete successfully", func() {
			res, err = dpdkClient.DeleteVirtualIP(ctx, res.InterfaceID)
			Expect(err).ToNot(HaveOccurred())

			res, err = dpdkClient.GetVirtualIP(ctx, "vm4")
			Expect(err).To(HaveOccurred())
		})
	})
})

var _ = Describe("loadbalancer", Label("loadbalancer"), func() {
	ctx := context.TODO()

	Context("When using loadbalancer functions", Ordered, func() {
		var res *api.LoadBalancer
		var err error

		It("should create successfully", func() {
			var lbVipIp = netip.MustParseAddr("10.20.30.40")
			lb := api.LoadBalancer{
				LoadBalancerMeta: api.LoadBalancerMeta{
					ID: "lb1",
				},
				Spec: api.LoadBalancerSpec{
					VNI:     100,
					LbVipIP: &lbVipIp,
					Lbports: []api.LBPort{
						{
							Protocol: 6,
							Port:     443,
						},
						{
							Protocol: 17,
							Port:     53,
						},
					},
				},
			}

			res, err = dpdkClient.CreateLoadBalancer(ctx, &lb)
			Expect(err).ToNot(HaveOccurred())

			Expect(res.ID).To(Equal("lb1"))
			Expect(res.Spec.VNI).To(Equal(uint32(100)))
		})

		It("should not be created when already existing", func() {
			res, err = dpdkClient.CreateLoadBalancer(ctx, res)
			Expect(err).To(HaveOccurred())

			Expect(res.Status.Code).To(Equal(uint32(errors.ALREADY_EXISTS)))
		})

		It("should get successfully", func() {
			res, err = dpdkClient.GetLoadBalancer(ctx, res.ID)
			Expect(err).ToNot(HaveOccurred())

			Expect(res.Spec.LbVipIP.String()).To(Equal("10.20.30.40"))
			Expect(res.Spec.Lbports[0].Port).To(Equal(uint32(443)))
		})

		It("should delete successfully", func() {
			res, err = dpdkClient.DeleteLoadBalancer(ctx, res.ID)
			Expect(err).ToNot(HaveOccurred())

			res, err = dpdkClient.GetLoadBalancer(ctx, "lb1")
			Expect(err).To(HaveOccurred())
			Expect(res.Status.Code).To(Equal(uint32(errors.NOT_FOUND)))
		})
	})
})
