// Copyright 2022 OnMetal authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package api

import (
	"fmt"
	"net/netip"
	"strconv"
	"strings"

	proto "github.com/onmetal/net-dpservice-go/proto"
)

func ProtoLoadBalancerToLoadBalancer(dpdkLB *proto.GetLoadBalancerResponse, lbID string) (*LoadBalancer, error) {

	var underlayRoute netip.Addr
	if underlayRouteString := string(dpdkLB.GetUnderlayRoute()); underlayRouteString != "" {
		var err error
		underlayRoute, err = netip.ParseAddr(string(dpdkLB.GetUnderlayRoute()))
		if err != nil {
			return nil, fmt.Errorf("error parsing underlay ip: %w", err)
		}
	}
	var lbip netip.Addr
	if lbipString := string(dpdkLB.GetLbVipIP().Address); lbipString != "" {
		var err error
		lbip, err = netip.ParseAddr(string(dpdkLB.GetLbVipIP().Address))
		if err != nil {
			return nil, fmt.Errorf("error parsing lb ip: %w", err)
		}
	}
	var lbports = make([]LBPort, 0, len(dpdkLB.Lbports))
	var p LBPort
	for _, lbport := range dpdkLB.Lbports {
		p.Protocol = uint32(lbport.Protocol)
		p.Port = lbport.Port
		lbports = append(lbports, p)
	}

	return &LoadBalancer{
		TypeMeta: TypeMeta{
			Kind: LoadBalancerKind,
		},
		LoadBalancerMeta: LoadBalancerMeta{
			ID: lbID,
		},
		Spec: LoadBalancerSpec{
			VNI:           dpdkLB.Vni,
			LbVipIP:       &lbip,
			Lbports:       lbports,
			UnderlayRoute: &underlayRoute,
		},
		Status: Status{
			Error:   dpdkLB.Status.Error,
			Message: dpdkLB.Status.Message,
		},
	}, nil
}

func LbipToProtoLbip(lbip netip.Addr) *proto.LBIP {
	return &proto.LBIP{IpVersion: NetIPAddrToProtoIPVersion(lbip), Address: []byte(lbip.String())}
}

func ProtoLbipToLbip(protolbip proto.LBIP) *netip.Addr {
	var ip netip.Addr
	if lbipString := string(protolbip.Address); lbipString != "" {
		var err error
		ip, err = netip.ParseAddr(string(protolbip.Address))
		if err != nil {
			return nil
		}
	}
	return &ip
}

func StringLbportToLbport(lbport string) (LBPort, error) {
	p := strings.Split(lbport, "/")
	protocolName := strings.ToLower(p[0])
	switch protocolName {
	case "icmp", "tcp", "udp", "sctp":
		protocolName = strings.ToUpper(protocolName)
	case "icmpv6":
		protocolName = "ICMPv6"
	default:
		return LBPort{}, fmt.Errorf("unsupported protocol")
	}
	protocol := proto.Protocol_value[protocolName]
	port, err := strconv.Atoi(p[1])
	if err != nil {
		return LBPort{}, fmt.Errorf("error parsing port number: %w", err)
	}
	return LBPort{Protocol: uint32(protocol), Port: uint32(port)}, nil
}

func ProtoInterfaceToInterface(dpdkIface *proto.Interface) (*Interface, error) {
	var ips []netip.Addr

	if ipv4String := string(dpdkIface.GetPrimaryIPv4Address()); ipv4String != "" {
		ip, err := netip.ParseAddr(ipv4String)
		if err != nil {
			return nil, fmt.Errorf("error parsing primary ipv4: %w", err)
		}

		ips = append(ips, ip)
	}

	if ipv6String := string(dpdkIface.GetPrimaryIPv6Address()); ipv6String != "" {
		ip, err := netip.ParseAddr(ipv6String)
		if err != nil {
			return nil, fmt.Errorf("error parsing primary ipv6: %w", err)
		}

		ips = append(ips, ip)
	}

	var underlayRoute netip.Addr
	if underlayRouteString := string(dpdkIface.GetUnderlayRoute()); underlayRouteString != "" {
		var err error
		underlayRoute, err = netip.ParseAddr(string(dpdkIface.GetUnderlayRoute()))
		if err != nil {
			return nil, fmt.Errorf("error parsing underlay ip: %w", err)
		}
	}

	return &Interface{
		TypeMeta: TypeMeta{
			Kind: InterfaceKind,
		},
		InterfaceMeta: InterfaceMeta{
			ID: string(dpdkIface.InterfaceID),
		},
		Spec: InterfaceSpec{
			VNI:           dpdkIface.GetVni(),
			Device:        dpdkIface.GetPciDpName(),
			IPs:           ips,
			UnderlayRoute: &underlayRoute,
		},
	}, nil
}

func NetIPAddrToProtoIPVersion(addr netip.Addr) proto.IPVersion {
	switch {
	case addr.Is4():
		return proto.IPVersion_IPv4
	case addr.Is6():
		return proto.IPVersion_IPv6
	default:
		return 0
	}
}

func NetIPAddrToProtoIPConfig(addr netip.Addr) *proto.IPConfig {
	if !addr.IsValid() {
		return nil
	}

	return &proto.IPConfig{
		IpVersion:      NetIPAddrToProtoIPVersion(addr),
		PrimaryAddress: []byte(addr.String()),
	}
}

func ProtoVirtualIPToVirtualIP(interfaceID string, dpdkVIP *proto.InterfaceVIPIP) (*VirtualIP, error) {
	ip, err := netip.ParseAddr(string(dpdkVIP.GetAddress()))
	if err != nil {
		return nil, fmt.Errorf("error parsing virtual ip address: %w", err)
	}

	underlayRoute, err := netip.ParseAddr(string(dpdkVIP.UnderlayRoute))
	if err != nil {
		return nil, fmt.Errorf("error parsing underlay route: %w", err)
	}

	return &VirtualIP{
		TypeMeta: TypeMeta{
			Kind: VirtualIPKind,
		},
		VirtualIPMeta: VirtualIPMeta{
			InterfaceID: interfaceID,
		},
		Spec: VirtualIPSpec{
			IP:            ip,
			UnderlayRoute: &underlayRoute,
		},
		Status: ProtoStatusToStatus(dpdkVIP.Status),
	}, nil
}

func ProtoPrefixToPrefix(interfaceID string, dpdkPrefix *proto.Prefix) (*Prefix, error) {
	addr, err := netip.ParseAddr(string(dpdkPrefix.GetAddress()))
	if err != nil {
		return nil, fmt.Errorf("error parsing dpdk prefix address: %w", err)
	}

	prefix := netip.PrefixFrom(addr, int(dpdkPrefix.GetPrefixLength()))

	underlayRoute, err := netip.ParseAddr(string(dpdkPrefix.UnderlayRoute))
	if err != nil {
		return nil, fmt.Errorf("error parsing underlay route: %w", err)
	}

	return &Prefix{
		TypeMeta: TypeMeta{
			Kind: PrefixKind,
		},
		PrefixMeta: PrefixMeta{
			InterfaceID: interfaceID,
		},
		Spec: PrefixSpec{
			Prefix:        prefix,
			UnderlayRoute: &underlayRoute,
		},
	}, nil
}

func ProtoRouteToRoute(vni uint32, dpdkRoute *proto.Route) (*Route, error) {
	prefixAddr, err := netip.ParseAddr(string(dpdkRoute.GetPrefix().GetAddress()))
	if err != nil {
		return nil, fmt.Errorf("error parsing prefix address: %w", err)
	}

	prefix := netip.PrefixFrom(prefixAddr, int(dpdkRoute.GetPrefix().GetPrefixLength()))

	nextHopIP, err := netip.ParseAddr(string(dpdkRoute.GetNexthopAddress()))
	if err != nil {
		return nil, fmt.Errorf("error parsing netxt hop address: %w", err)
	}

	return &Route{
		TypeMeta: TypeMeta{
			RouteKind,
		},
		RouteMeta: RouteMeta{
			VNI: vni,
		},
		Spec: RouteSpec{Prefix: &prefix,
			NextHop: &RouteNextHop{
				VNI: dpdkRoute.GetNexthopVNI(),
				IP:  &nextHopIP,
			}},
	}, nil
}

func ProtoLBPrefixToProtoPrefix(lbprefix proto.LBPrefix) *proto.Prefix {
	return &proto.Prefix{
		IpVersion:     lbprefix.IpVersion,
		Address:       lbprefix.Address,
		PrefixLength:  lbprefix.PrefixLength,
		UnderlayRoute: lbprefix.UnderlayRoute,
	}
}

func ProtoNatToNat(dpdkNat *proto.GetNATResponse, interfaceID string) (*Nat, error) {
	var underlayRoute netip.Addr
	if underlayRouteString := string(dpdkNat.GetUnderlayRoute()); underlayRouteString != "" {
		var err error
		underlayRoute, err = netip.ParseAddr(string(dpdkNat.GetUnderlayRoute()))
		if err != nil {
			return nil, fmt.Errorf("error parsing underlay ip: %w", err)
		}
	}
	var natvipip netip.Addr
	if natvipipString := string(dpdkNat.GetNatVIPIP().Address); natvipipString != "" {
		var err error
		natvipip, err = netip.ParseAddr(string(dpdkNat.GetNatVIPIP().Address))
		if err != nil {
			return nil, fmt.Errorf("error parsing nat ip: %w", err)
		}
	}

	return &Nat{
		TypeMeta: TypeMeta{
			Kind: NatKind,
		},
		NatMeta: NatMeta{
			InterfaceID: interfaceID,
		},
		Spec: NatSpec{
			NatVIPIP:      &natvipip,
			MinPort:       dpdkNat.MinPort,
			MaxPort:       dpdkNat.MaxPort,
			UnderlayRoute: &underlayRoute,
		},
		Status: Status{
			Error:   dpdkNat.Status.Error,
			Message: dpdkNat.Status.Message,
		},
	}, nil
}

func ProtoFwRuleToFwRule(dpdkFwRule *proto.FirewallRule, interfaceID string) (*FirewallRule, error) {

	srcPrefix, err := netip.ParsePrefix(string(dpdkFwRule.SourcePrefix.Address) + "/" + strconv.Itoa(int(dpdkFwRule.SourcePrefix.PrefixLength)))
	if err != nil {
		return nil, fmt.Errorf("error converting prefix: %w", err)
	}

	dstPrefix, err := netip.ParsePrefix(string(dpdkFwRule.DestinationPrefix.Address) + "/" + strconv.Itoa(int(dpdkFwRule.DestinationPrefix.PrefixLength)))
	if err != nil {
		return nil, fmt.Errorf("error converting prefix: %w", err)
	}
	var direction, action, ipv string
	if dpdkFwRule.Direction == 0 {
		direction = "Ingress"
	} else {
		direction = "Egress"
	}
	if dpdkFwRule.Action == 0 {
		action = "Drop"
	} else {
		action = "Accept"
	}
	if dpdkFwRule.IpVersion == 0 {
		ipv = "IPv4"
	} else {
		ipv = "IPv6"
	}

	return &FirewallRule{
		TypeMeta: TypeMeta{Kind: FirewallRuleKind},
		FirewallRuleMeta: FirewallRuleMeta{
			InterfaceID: interfaceID,
		},
		Spec: FirewallRuleSpec{
			RuleID:            string(dpdkFwRule.RuleID),
			TrafficDirection:  direction,
			FirewallAction:    action,
			Priority:          dpdkFwRule.Priority,
			IpVersion:         ipv,
			SourcePrefix:      &srcPrefix,
			DestinationPrefix: &dstPrefix,
			ProtocolFilter:    dpdkFwRule.ProtocolFilter,
		},
	}, nil
}

func ProtoStatusToStatus(dpdkStatus *proto.Status) Status {
	return Status{
		Error:   dpdkStatus.Error,
		Message: dpdkStatus.Message,
	}
}
