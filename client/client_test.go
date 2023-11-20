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
	dpdkproto "github.com/onmetal/net-dpservice-go/proto"
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

			vni, err := dpdkClient.GetVni(ctx, 200, 0)
			Expect(err).ToNot(HaveOccurred())

			Expect(vni.Spec.InUse).To(BeFalse())

			res, err = dpdkClient.CreateInterface(ctx, &iface)
			Expect(err).ToNot(HaveOccurred())

			Expect(res.ID).To(Equal("vm4"))
			Expect(res.Spec.VNI).To(Equal(uint32(200)))

			vni, err = dpdkClient.GetVni(ctx, 200, 0)
			Expect(err).ToNot(HaveOccurred())

			Expect(vni.Spec.InUse).To(BeTrue())
		})

		It("should not be created when already existing", func() {
			res, err := dpdkClient.CreateInterface(ctx, res)
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

var _ = Describe("interface related", func() {
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

	Context("When using prefix functions", Label("prefix"), Ordered, func() {
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
			res, err := dpdkClient.CreatePrefix(ctx, res)
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

	Context("When using lbprefix functions", Label("lbprefix"), Ordered, func() {
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
			res, err := dpdkClient.CreateLoadBalancerPrefix(ctx, res)
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

	Context("When using virtualIP functions", Label("vip"), Ordered, func() {
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
			res, err := dpdkClient.CreateVirtualIP(ctx, res)
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

	Context("When using nat functions", Label("nat"), Ordered, func() {
		var res *api.Nat
		var err error

		It("should create successfully", func() {
			ip := netip.MustParseAddr("10.20.30.40")
			nat := api.Nat{
				NatMeta: api.NatMeta{
					InterfaceID: "vm4",
				},
				Spec: api.NatSpec{
					NatIP:   &ip,
					MinPort: 30000,
					MaxPort: 30100,
				},
			}

			res, err = dpdkClient.CreateNat(ctx, &nat)
			Expect(err).ToNot(HaveOccurred())

			Expect(res.InterfaceID).To(Equal("vm4"))
			Expect(res.Spec.NatIP.String()).To(Equal("10.20.30.40"))
		})

		It("should not be created when already existing", func() {
			res, err := dpdkClient.CreateNat(ctx, res)
			Expect(err).To(HaveOccurred())

			Expect(res.Status.Code).To(Equal(uint32(errors.SNAT_EXISTS)))
		})

		It("should get successfully", func() {
			res, err = dpdkClient.GetNat(ctx, "vm4")
			Expect(err).ToNot(HaveOccurred())

			Expect(res.InterfaceID).To(Equal("vm4"))
			Expect(res.Spec.UnderlayRoute).ToNot(BeNil())
			Expect(res.Spec.MinPort).To(Equal(uint32(30000)))
		})

		It("should list localNats successfully", func() {
			localNats, err := dpdkClient.ListLocalNats(ctx, res.Spec.NatIP)
			Expect(err).ToNot(HaveOccurred())

			Expect(len(localNats.Items)).To(Equal(1))
			Expect(localNats.Items[0].Kind).To(Equal(api.NatKind))
			Expect(localNats.Items[0].Spec.MinPort).To(Equal(uint32(30000)))
		})

		It("should delete successfully", func() {
			res, err = dpdkClient.DeleteNat(ctx, res.InterfaceID)
			Expect(err).ToNot(HaveOccurred())

			res, err = dpdkClient.GetNat(ctx, "vm4")
			Expect(err).To(HaveOccurred())
		})
	})

	Context("When using neighbor nat functions", Label("neighbornat"), Ordered, func() {
		var res *api.NeighborNat
		var err error

		It("should create successfully", func() {
			natIp := netip.MustParseAddr("10.20.30.40")
			underlayRoute := netip.MustParseAddr("ff80::1")
			neighborNat := api.NeighborNat{
				NeighborNatMeta: api.NeighborNatMeta{
					NatIP: &natIp,
				},
				Spec: api.NeighborNatSpec{
					Vni:           100,
					MinPort:       30000,
					MaxPort:       30100,
					UnderlayRoute: &underlayRoute,
				},
			}

			res, err = dpdkClient.CreateNeighborNat(ctx, &neighborNat)
			Expect(err).ToNot(HaveOccurred())

			Expect(res.NatIP.String()).To(Equal("10.20.30.40"))
			Expect(res.Spec.Vni).To(Equal(uint32(100)))
		})

		It("should not be created when already existing", func() {
			res, err := dpdkClient.CreateNeighborNat(ctx, res)
			Expect(err).To(HaveOccurred())

			Expect(res.Status.Code).To(Equal(uint32(errors.ALREADY_EXISTS)))
		})

		It("should list successfully", func() {
			neighborNats, err := dpdkClient.ListNeighborNats(ctx, res.NatIP)
			Expect(err).ToNot(HaveOccurred())

			Expect(len(neighborNats.Items)).To(Equal(1))
			// TODO: items kind should be NeighborNat
			Expect(neighborNats.Items[0].Kind).To(Equal(api.NatKind))
			Expect(neighborNats.Items[0].Spec.MinPort).To(Equal(uint32(30000)))
		})

		It("should list Nats successfully", func() {
			nats, err := dpdkClient.ListNats(ctx, res.NatIP, "any")
			Expect(err).ToNot(HaveOccurred())

			Expect(len(nats.Items)).To(Equal(1))
			Expect(nats.Items[0].Spec.MinPort).To(Equal(uint32(30000)))
		})

		It("should delete successfully", func() {
			res, err = dpdkClient.DeleteNeighborNat(ctx, res)
			Expect(err).ToNot(HaveOccurred())

			neighborNats, err := dpdkClient.ListNeighborNats(ctx, res.NatIP)
			Expect(err).ToNot(HaveOccurred())
			Expect(len(neighborNats.Items)).To(Equal(0))
		})
	})

	Context("When using route functions", Label("route"), Ordered, func() {
		var res *api.Route
		var err error

		It("should create successfully", func() {
			prefix := netip.MustParsePrefix("10.100.3.0/24")
			nextHopIp := netip.MustParseAddr("fc00:2::64:0:1")
			route := api.Route{
				RouteMeta: api.RouteMeta{
					VNI: 200,
				},
				Spec: api.RouteSpec{
					Prefix: &prefix,
					NextHop: &api.RouteNextHop{
						VNI: 0,
						IP:  &nextHopIp,
					},
				},
			}
			res, err = dpdkClient.CreateRoute(ctx, &route)
			Expect(err).ToNot(HaveOccurred())

			Expect(res.VNI).To(Equal(uint32(200)))
			Expect(res.Spec.Prefix.String()).To(Equal("10.100.3.0/24"))
		})

		It("should not be created when already existing", func() {
			res, err := dpdkClient.CreateRoute(ctx, res)
			Expect(err).To(HaveOccurred())

			Expect(res.Status.Code).To(Equal(uint32(errors.ROUTE_EXISTS)))
		})

		It("should list successfully", func() {
			routes, err := dpdkClient.ListRoutes(ctx, 200)
			Expect(err).ToNot(HaveOccurred())

			Expect(len(routes.Items)).To(Equal(1))
			Expect(routes.Items[0].Kind).To(Equal(api.RouteKind))
		})

		It("should delete successfully", func() {
			res, err = dpdkClient.DeleteRoute(ctx, res.VNI, res.Spec.Prefix)
			Expect(err).ToNot(HaveOccurred())

			routes, err := dpdkClient.ListRoutes(ctx, 200)
			Expect(err).ToNot(HaveOccurred())
			Expect(len(routes.Items)).To(Equal(0))
		})
	})

	Context("When using firewall rule functions", Label("fwrule"), Ordered, func() {
		var res *api.FirewallRule
		var err error

		It("should create successfully", func() {
			src := netip.MustParsePrefix("1.1.1.1/32")
			dst := netip.MustParsePrefix("5.5.5.0/24")
			fwRule := api.FirewallRule{
				FirewallRuleMeta: api.FirewallRuleMeta{
					InterfaceID: "vm4",
				},
				Spec: api.FirewallRuleSpec{
					RuleID:            "Rule1",
					TrafficDirection:  "ingress",
					FirewallAction:    "accept",
					Priority:          1000,
					SourcePrefix:      &src,
					DestinationPrefix: &dst,
					ProtocolFilter: &dpdkproto.ProtocolFilter{
						Filter: &dpdkproto.ProtocolFilter_Tcp{
							Tcp: &dpdkproto.TcpFilter{
								SrcPortLower: 1,
								SrcPortUpper: 65535,
								DstPortLower: 500,
								DstPortUpper: 600,
							},
						},
					},
				},
			}

			res, err = dpdkClient.CreateFirewallRule(ctx, &fwRule)
			Expect(err).ToNot(HaveOccurred())

			Expect(res.InterfaceID).To(Equal("vm4"))
			Expect(res.Spec.RuleID).To(Equal("Rule1"))
		})

		It("should not be created when already existing", func() {
			res, err := dpdkClient.CreateFirewallRule(ctx, res)
			Expect(err).To(HaveOccurred())

			Expect(res.Status.Code).To(Equal(uint32(errors.ALREADY_EXISTS)))
		})

		It("should get successfully", func() {
			res, err = dpdkClient.GetFirewallRule(ctx, res.InterfaceID, "Rule1")
			Expect(err).ToNot(HaveOccurred())

			Expect(res.Spec.TrafficDirection).To(Equal("Ingress"))
			Expect(res.Spec.SourcePrefix.String()).To(Equal("1.1.1.1/32"))
		})

		It("should list successfully", func() {
			fwRules, err := dpdkClient.ListFirewallRules(ctx, res.InterfaceID)
			Expect(err).ToNot(HaveOccurred())

			Expect(len(fwRules.Items)).To(Equal(1))
			Expect(fwRules.Items[0].Kind).To(Equal(api.FirewallRuleKind))
			Expect(fwRules.Items[0].Spec.Priority).To(Equal(uint32(1000)))
		})

		It("should delete successfully", func() {
			res, err = dpdkClient.DeleteFirewallRule(ctx, res.InterfaceID, "Rule1")
			Expect(err).ToNot(HaveOccurred())

			res, err = dpdkClient.GetFirewallRule(ctx, res.InterfaceID, "Rule1")
			Expect(err).To(HaveOccurred())
			Expect(res.Status.Code).To(Equal(uint32(errors.NOT_FOUND)))
		})
	})
})

var _ = Describe("loadbalancer related", func() {
	ctx := context.TODO()

	Context("When using loadbalancer functions", Label("loadbalancer"), Ordered, func() {
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
			res, err := dpdkClient.CreateLoadBalancer(ctx, res)
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

	Context("When using loadbalancer target functions", Label("lbtarget"), Ordered, func() {
		var res *api.LoadBalancerTarget
		var lb api.LoadBalancer
		var err error

		It("should create successfully", func() {
			var lbVipIp = netip.MustParseAddr("10.20.30.40")
			lb = api.LoadBalancer{
				LoadBalancerMeta: api.LoadBalancerMeta{
					ID: "lb2",
				},
				Spec: api.LoadBalancerSpec{
					VNI:     200,
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

			_, err = dpdkClient.CreateLoadBalancer(ctx, &lb)
			Expect(err).ToNot(HaveOccurred())

			targetIp := netip.MustParseAddr("ff80::5")
			lbtarget := api.LoadBalancerTarget{
				LoadBalancerTargetMeta: api.LoadBalancerTargetMeta{
					LoadbalancerID: "lb2",
				},
				Spec: api.LoadBalancerTargetSpec{
					TargetIP: &targetIp,
				},
			}

			res, err = dpdkClient.CreateLoadBalancerTarget(ctx, &lbtarget)
			Expect(err).ToNot(HaveOccurred())

			Expect(res.LoadbalancerID).To(Equal("lb2"))
			Expect(res.Spec.TargetIP.String()).To(Equal("ff80::5"))
		})

		It("should not be created when already existing", func() {
			res, err := dpdkClient.CreateLoadBalancerTarget(ctx, res)
			Expect(err).To(HaveOccurred())

			Expect(res.Status.Code).To(Equal(uint32(errors.ALREADY_EXISTS)))
		})

		It("should list successfully", func() {
			lbtargets, err := dpdkClient.ListLoadBalancerTargets(ctx, res.LoadbalancerID)
			Expect(err).ToNot(HaveOccurred())

			Expect(lbtargets.Items[0].LoadbalancerID).To(Equal("lb2"))
			Expect(len(lbtargets.Items)).To(Equal(1))
			Expect(lbtargets.Items[0].Kind).To(Equal("LoadBalancerTarget"))
		})

		It("should delete successfully", func() {
			res, err = dpdkClient.DeleteLoadBalancerTarget(ctx, res.LoadbalancerID, res.Spec.TargetIP)
			Expect(err).ToNot(HaveOccurred())

			lbtargets, err := dpdkClient.ListLoadBalancerTargets(ctx, "lb2")
			Expect(err).ToNot(HaveOccurred())
			Expect(len(lbtargets.Items)).To(Equal(0))

			_, err = dpdkClient.DeleteLoadBalancer(ctx, "lb2")
			Expect(err).ToNot(HaveOccurred())
		})
	})
})

var _ = Describe("init", Label("init"), func() {
	ctx := context.TODO()

	Context("When using init functions", Ordered, func() {
		var res *api.Initialized
		var err error

		It("should initialize successfully", func() {
			init, err := dpdkClient.Initialize(ctx)
			Expect(err).ToNot(HaveOccurred())

			Expect(init.Spec.UUID).ToNot(Equal(""))

			// Initializing again should return same UUID
			res, err = dpdkClient.Initialize(ctx)
			Expect(err).ToNot(HaveOccurred())

			Expect(init.Spec.UUID).To(Equal(res.Spec.UUID))
		})

		It("should check if initialized successfully", func() {
			res, err = dpdkClient.CheckInitialized(ctx)
			Expect(err).ToNot(HaveOccurred())

			Expect(res.Spec.UUID).ToNot(Equal(""))
		})

		It("should get version successfully", func() {
			clientVersion := api.Version{
				VersionMeta: api.VersionMeta{
					ClientProtocol: "0.0.1",
					ClientName:     "testClient",
					ClientVersion:  "0.0.1"},
			}
			version, err := dpdkClient.GetVersion(ctx, &clientVersion)
			Expect(err).ToNot(HaveOccurred())

			Expect(version.ClientName).To(Equal("testClient"))
			Expect(version.Spec.ServiceProtocol).ToNot(Equal(""))
			Expect(version.Spec.ServiceVersion).ToNot(Equal(""))
		})
	})
})

// TODO: add capture functions tests
// TODO: add negstive tests
