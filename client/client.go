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

package client

import (
	"context"
	"fmt"
	"net/netip"
	"strings"

	"github.com/onmetal/net-dpservice-go/api"
	"github.com/onmetal/net-dpservice-go/errors"
	"github.com/onmetal/net-dpservice-go/netiputil"
	dpdkproto "github.com/onmetal/net-dpservice-go/proto"
)

type Client interface {
	GetLoadBalancer(ctx context.Context, id string) (*api.LoadBalancer, error)
	CreateLoadBalancer(ctx context.Context, lb *api.LoadBalancer) (*api.LoadBalancer, error)
	DeleteLoadBalancer(ctx context.Context, id string) (*api.LoadBalancer, error)

	ListLoadBalancerPrefixes(ctx context.Context, interfaceID string) (*api.PrefixList, error)
	CreateLoadBalancerPrefix(ctx context.Context, prefix *api.LoadBalancerPrefix) (*api.LoadBalancerPrefix, error)
	DeleteLoadBalancerPrefix(ctx context.Context, interfaceID string, prefix netip.Prefix) (*api.LoadBalancerPrefix, error)

	GetLoadBalancerTargets(ctx context.Context, interfaceID string) (*api.LoadBalancerTargetList, error)
	AddLoadBalancerTarget(ctx context.Context, lbtarget *api.LoadBalancerTarget) (*api.LoadBalancerTarget, error)
	DeleteLoadBalancerTarget(ctx context.Context, id string, targetIP netip.Addr) (*api.LoadBalancerTarget, error)

	GetInterface(ctx context.Context, id string) (*api.Interface, error)
	ListInterfaces(ctx context.Context) (*api.InterfaceList, error)
	CreateInterface(ctx context.Context, iface *api.Interface) (*api.Interface, error)
	DeleteInterface(ctx context.Context, id string) (*api.Interface, error)

	GetVirtualIP(ctx context.Context, interfaceID string) (*api.VirtualIP, error)
	AddVirtualIP(ctx context.Context, virtualIP *api.VirtualIP) (*api.VirtualIP, error)
	DeleteVirtualIP(ctx context.Context, interfaceID string) (*api.VirtualIP, error)

	ListPrefixes(ctx context.Context, interfaceID string) (*api.PrefixList, error)
	AddPrefix(ctx context.Context, prefix *api.Prefix) (*api.Prefix, error)
	DeletePrefix(ctx context.Context, interfaceID string, prefix netip.Prefix) (*api.Prefix, error)

	ListRoutes(ctx context.Context, vni uint32) (*api.RouteList, error)
	AddRoute(ctx context.Context, route *api.Route) (*api.Route, error)
	DeleteRoute(ctx context.Context, vni uint32, prefix netip.Prefix) (*api.Route, error)

	GetNat(ctx context.Context, interfaceID string) (*api.Nat, error)
	AddNat(ctx context.Context, nat *api.Nat) (*api.Nat, error)
	DeleteNat(ctx context.Context, interfaceID string) (*api.Nat, error)

	AddNeighborNat(ctx context.Context, nat *api.NeighborNat) (*api.NeighborNat, error)
	GetNATInfo(ctx context.Context, natVIPIP netip.Addr, natType string) (*api.NatList, error)
	DeleteNeighborNat(ctx context.Context, neigbhorNat api.NeighborNat) (*api.NeighborNat, error)

	ListFirewallRules(ctx context.Context, interfaceID string) (*api.FirewallRuleList, error)
	AddFirewallRule(ctx context.Context, fwRule *api.FirewallRule) (*api.FirewallRule, error)
	GetFirewallRule(ctx context.Context, interfaceID string, ruleID string) (*api.FirewallRule, error)
	DeleteFirewallRule(ctx context.Context, interfaceID string, ruleID string) (*api.FirewallRule, error)

	Initialized(ctx context.Context) (string, error)
	Init(ctx context.Context, initConfig dpdkproto.InitConfig) (*api.Init, error)
	GetVni(ctx context.Context, vni uint32, vniType uint8) (*api.Vni, error)
	ResetVni(ctx context.Context, vni uint32, vniType uint8) (*api.Vni, error)
}

type client struct {
	dpdkproto.DPDKonmetalClient
}

func NewClient(protoClient dpdkproto.DPDKonmetalClient) Client {
	return &client{protoClient}
}

func (c *client) GetLoadBalancer(ctx context.Context, id string) (*api.LoadBalancer, error) {
	res, err := c.DPDKonmetalClient.GetLoadBalancer(ctx, &dpdkproto.GetLoadBalancerRequest{LoadBalancerID: []byte(id)})
	if err != nil {
		return &api.LoadBalancer{}, err
	}
	retLoadBalancer := &api.LoadBalancer{
		TypeMeta:         api.TypeMeta{Kind: api.LoadBalancerKind},
		LoadBalancerMeta: api.LoadBalancerMeta{ID: id},
		Status:           api.ProtoStatusToStatus(res.Status),
	}
	if errorCode := res.GetStatus().GetError(); errorCode != 0 {
		return retLoadBalancer, errors.ErrServerError
	}
	return api.ProtoLoadBalancerToLoadBalancer(res, id)
}

func (c *client) CreateLoadBalancer(ctx context.Context, lb *api.LoadBalancer) (*api.LoadBalancer, error) {
	var lbPorts = make([]*dpdkproto.LBPort, 0, len(lb.Spec.Lbports))
	for _, p := range lb.Spec.Lbports {
		lbPort := &dpdkproto.LBPort{Port: p.Port, Protocol: dpdkproto.Protocol(p.Protocol)}
		lbPorts = append(lbPorts, lbPort)
	}
	res, err := c.DPDKonmetalClient.CreateLoadBalancer(ctx, &dpdkproto.CreateLoadBalancerRequest{
		LoadBalancerID: []byte(lb.LoadBalancerMeta.ID),
		Vni:            lb.Spec.VNI,
		LbVipIP:        api.LbipToProtoLbip(*lb.Spec.LbVipIP),
		Lbports:        lbPorts,
	})
	if err != nil {
		return &api.LoadBalancer{}, err
	}
	retLoadBalancer := &api.LoadBalancer{
		TypeMeta:         api.TypeMeta{Kind: api.LoadBalancerKind},
		LoadBalancerMeta: lb.LoadBalancerMeta,
		Status:           api.ProtoStatusToStatus(res.Status),
	}
	if errorCode := res.GetStatus().GetError(); errorCode != 0 {
		return retLoadBalancer, errors.ErrServerError
	}

	underlayRoute, err := netip.ParseAddr(string(res.GetUnderlayRoute()))
	if err != nil {
		return retLoadBalancer, fmt.Errorf("error parsing underlay route: %w", err)
	}
	retLoadBalancer.Spec = lb.Spec
	retLoadBalancer.Spec.UnderlayRoute = &underlayRoute

	return retLoadBalancer, nil
}

func (c *client) DeleteLoadBalancer(ctx context.Context, id string) (*api.LoadBalancer, error) {
	res, err := c.DPDKonmetalClient.DeleteLoadBalancer(ctx, &dpdkproto.DeleteLoadBalancerRequest{LoadBalancerID: []byte(id)})
	if err != nil {
		return &api.LoadBalancer{}, err
	}
	retLoadBalancer := &api.LoadBalancer{
		TypeMeta:         api.TypeMeta{Kind: api.LoadBalancerKind},
		LoadBalancerMeta: api.LoadBalancerMeta{ID: id},
		Status:           api.ProtoStatusToStatus(res),
	}
	if errorCode := res.GetError(); errorCode != 0 {
		return retLoadBalancer, errors.ErrServerError
	}
	return retLoadBalancer, nil
}

func (c *client) ListLoadBalancerPrefixes(ctx context.Context, interfaceID string) (*api.PrefixList, error) {
	res, err := c.DPDKonmetalClient.ListInterfaceLoadBalancerPrefixes(ctx, &dpdkproto.ListInterfaceLoadBalancerPrefixesRequest{
		InterfaceID: []byte(interfaceID),
	})
	if err != nil {
		return nil, err
	}

	prefixes := make([]api.Prefix, len(res.GetPrefixes()))
	for i, dpdkPrefix := range res.GetPrefixes() {
		prefix, err := api.ProtoPrefixToPrefix(interfaceID, api.ProtoLBPrefixToProtoPrefix(*dpdkPrefix))
		prefix.Kind = api.LoadBalancerPrefixKind
		if err != nil {
			return nil, err
		}

		prefixes[i] = *prefix
	}

	return &api.PrefixList{
		TypeMeta:       api.TypeMeta{Kind: "LoadBalancerPrefixList"},
		PrefixListMeta: api.PrefixListMeta{InterfaceID: interfaceID},
		Items:          prefixes,
	}, nil
}

func (c *client) CreateLoadBalancerPrefix(ctx context.Context, lbprefix *api.LoadBalancerPrefix) (*api.LoadBalancerPrefix, error) {
	res, err := c.DPDKonmetalClient.CreateInterfaceLoadBalancerPrefix(ctx, &dpdkproto.CreateInterfaceLoadBalancerPrefixRequest{
		InterfaceID: &dpdkproto.InterfaceIDMsg{
			InterfaceID: []byte(lbprefix.InterfaceID),
		},
		Prefix: &dpdkproto.Prefix{
			IpVersion:    api.NetIPAddrToProtoIPVersion(lbprefix.Spec.Prefix.Addr()),
			Address:      []byte(lbprefix.Spec.Prefix.Addr().String()),
			PrefixLength: uint32(lbprefix.Spec.Prefix.Bits()),
		},
	})
	if err != nil {
		return &api.LoadBalancerPrefix{}, err
	}
	retLBPrefix := &api.LoadBalancerPrefix{
		TypeMeta:               api.TypeMeta{Kind: api.LoadBalancerPrefixKind},
		LoadBalancerPrefixMeta: lbprefix.LoadBalancerPrefixMeta,
		Spec: api.LoadBalancerPrefixSpec{
			Prefix: lbprefix.Spec.Prefix,
		},
		Status: api.ProtoStatusToStatus(res.Status),
	}
	if errorCode := res.GetStatus().GetError(); errorCode != 0 {
		return retLBPrefix, errors.ErrServerError
	}
	underlayRoute, err := netip.ParseAddr(string(res.GetUnderlayRoute()))
	if err != nil {
		return retLBPrefix, fmt.Errorf("error parsing underlay route: %w", err)
	}
	retLBPrefix.Spec.UnderlayRoute = &underlayRoute
	return retLBPrefix, nil
}

func (c *client) DeleteLoadBalancerPrefix(ctx context.Context, interfaceID string, prefix netip.Prefix) (*api.LoadBalancerPrefix, error) {
	res, err := c.DPDKonmetalClient.DeleteInterfaceLoadBalancerPrefix(ctx, &dpdkproto.DeleteInterfaceLoadBalancerPrefixRequest{
		InterfaceID: &dpdkproto.InterfaceIDMsg{
			InterfaceID: []byte(interfaceID),
		},
		Prefix: &dpdkproto.Prefix{
			IpVersion:    api.NetIPAddrToProtoIPVersion(prefix.Addr()),
			Address:      []byte(prefix.Addr().String()),
			PrefixLength: uint32(prefix.Bits()),
		},
	})
	if err != nil {
		return &api.LoadBalancerPrefix{}, err
	}
	retLBPrefix := &api.LoadBalancerPrefix{
		TypeMeta:               api.TypeMeta{Kind: api.LoadBalancerPrefixKind},
		LoadBalancerPrefixMeta: api.LoadBalancerPrefixMeta{InterfaceID: interfaceID},
		Spec:                   api.LoadBalancerPrefixSpec{Prefix: prefix},
		Status:                 api.ProtoStatusToStatus(res),
	}
	if errorCode := res.GetError(); errorCode != 0 {
		return retLBPrefix, errors.ErrServerError
	}
	return retLBPrefix, nil
}

func (c *client) GetLoadBalancerTargets(ctx context.Context, loadBalancerID string) (*api.LoadBalancerTargetList, error) {
	res, err := c.DPDKonmetalClient.GetLoadBalancerTargets(ctx, &dpdkproto.GetLoadBalancerTargetsRequest{
		LoadBalancerID: []byte(loadBalancerID),
	})
	if err != nil {
		return &api.LoadBalancerTargetList{}, err
	}
	if errorCode := res.GetStatus().GetError(); errorCode != 0 {
		return &api.LoadBalancerTargetList{
			TypeMeta: api.TypeMeta{Kind: api.LoadBalancerTargetListKind},
			Status:   api.ProtoStatusToStatus(res.Status)}, errors.ErrServerError
	}

	lbtargets := make([]api.LoadBalancerTarget, len(res.GetTargetIPs()))
	for i, dpdkLBtarget := range res.GetTargetIPs() {
		var lbtarget api.LoadBalancerTarget
		lbtarget.TypeMeta.Kind = api.LoadBalancerTargetKind
		lbtarget.Spec.TargetIP = api.ProtoLbipToLbip(*dpdkLBtarget)
		lbtarget.LoadBalancerTargetMeta.LoadbalancerID = loadBalancerID

		lbtargets[i] = lbtarget
	}

	return &api.LoadBalancerTargetList{
		TypeMeta:                   api.TypeMeta{Kind: api.LoadBalancerTargetListKind},
		LoadBalancerTargetListMeta: api.LoadBalancerTargetListMeta{LoadBalancerID: loadBalancerID},
		Items:                      lbtargets,
		// TODO server is not returning correct status
		//Status:   api.ProtoStatusToStatus(res.Status),
	}, nil
}

func (c *client) AddLoadBalancerTarget(ctx context.Context, lbtarget *api.LoadBalancerTarget) (*api.LoadBalancerTarget, error) {
	res, err := c.DPDKonmetalClient.AddLoadBalancerTarget(ctx, &dpdkproto.AddLoadBalancerTargetRequest{
		LoadBalancerID: []byte(lbtarget.LoadBalancerTargetMeta.LoadbalancerID),
		TargetIP:       api.LbipToProtoLbip(*lbtarget.Spec.TargetIP),
	})
	if err != nil {
		return &api.LoadBalancerTarget{}, err
	}
	retLBTarget := &api.LoadBalancerTarget{
		TypeMeta:               api.TypeMeta{Kind: api.LoadBalancerTargetKind},
		LoadBalancerTargetMeta: lbtarget.LoadBalancerTargetMeta,
		Status:                 api.ProtoStatusToStatus(res),
	}
	if errorCode := res.GetError(); errorCode != 0 {
		return retLBTarget, errors.ErrServerError
	}
	retLBTarget.Spec = lbtarget.Spec
	return retLBTarget, nil
}

func (c *client) DeleteLoadBalancerTarget(ctx context.Context, lbid string, targetIP netip.Addr) (*api.LoadBalancerTarget, error) {
	res, err := c.DPDKonmetalClient.DeleteLoadBalancerTarget(ctx, &dpdkproto.DeleteLoadBalancerTargetRequest{
		LoadBalancerID: []byte(lbid),
		TargetIP:       api.LbipToProtoLbip(targetIP),
	})
	if err != nil {
		return &api.LoadBalancerTarget{}, err
	}
	retLBTarget := &api.LoadBalancerTarget{
		TypeMeta:               api.TypeMeta{Kind: api.LoadBalancerTargetKind},
		LoadBalancerTargetMeta: api.LoadBalancerTargetMeta{LoadbalancerID: lbid},
		Status:                 api.ProtoStatusToStatus(res),
	}
	if errorCode := res.GetError(); errorCode != 0 {
		return retLBTarget, errors.ErrServerError
	}
	return retLBTarget, nil
}

func (c *client) GetInterface(ctx context.Context, id string) (*api.Interface, error) {
	res, err := c.DPDKonmetalClient.GetInterface(ctx, &dpdkproto.InterfaceIDMsg{InterfaceID: []byte(id)})
	if err != nil {
		return &api.Interface{}, err
	}
	if errorCode := res.GetStatus().GetError(); errorCode != 0 {
		return &api.Interface{
			TypeMeta:      api.TypeMeta{Kind: api.InterfaceKind},
			InterfaceMeta: api.InterfaceMeta{ID: id},
			Status:        api.ProtoStatusToStatus(res.Status)}, errors.ErrServerError
	}
	return api.ProtoInterfaceToInterface(res.GetInterface())
}

func (c *client) ListInterfaces(ctx context.Context) (*api.InterfaceList, error) {
	res, err := c.DPDKonmetalClient.ListInterfaces(ctx, &dpdkproto.Empty{})
	if err != nil {
		return nil, err
	}

	ifaces := make([]api.Interface, len(res.GetInterfaces()))
	for i, dpdkIface := range res.GetInterfaces() {
		iface, err := api.ProtoInterfaceToInterface(dpdkIface)
		if err != nil {
			return nil, err
		}

		ifaces[i] = *iface
	}

	return &api.InterfaceList{
		TypeMeta: api.TypeMeta{Kind: api.InterfaceListKind},
		Items:    ifaces,
	}, nil
}

func (c *client) CreateInterface(ctx context.Context, iface *api.Interface) (*api.Interface, error) {
	req := dpdkproto.CreateInterfaceRequest{
		InterfaceType: dpdkproto.InterfaceType_VirtualInterface,
		InterfaceID:   []byte(iface.ID),
		Vni:           iface.Spec.VNI,
		Ipv4Config:    api.NetIPAddrToProtoIPConfig(netiputil.FindIPv4(iface.Spec.IPs)),
		Ipv6Config:    api.NetIPAddrToProtoIPConfig(netiputil.FindIPv6(iface.Spec.IPs)),
		DeviceName:    iface.Spec.Device,
	}
	if iface.Spec.PXE.FileName != "" && iface.Spec.PXE.Server != "" {
		req.Ipv4Config.PxeConfig = &dpdkproto.PXEConfig{NextServer: iface.Spec.PXE.Server, BootFileName: iface.Spec.PXE.FileName}
		req.Ipv6Config.PxeConfig = &dpdkproto.PXEConfig{NextServer: iface.Spec.PXE.Server, BootFileName: iface.Spec.PXE.FileName}
	}

	res, err := c.DPDKonmetalClient.CreateInterface(ctx, &req)
	if err != nil {
		return &api.Interface{}, err
	}
	retInterface := &api.Interface{
		TypeMeta:      iface.TypeMeta,
		InterfaceMeta: iface.InterfaceMeta,
		Status:        api.ProtoStatusToStatus(res.Response.Status),
	}
	if errorCode := res.GetResponse().GetStatus().GetError(); errorCode != 0 {
		return retInterface, errors.ErrServerError
	}

	underlayRoute, err := netip.ParseAddr(string(res.GetResponse().GetUnderlayRoute()))
	if err != nil {
		return retInterface, fmt.Errorf("error parsing underlay route: %w", err)
	}

	return &api.Interface{
		TypeMeta:      api.TypeMeta{Kind: api.InterfaceKind},
		InterfaceMeta: iface.InterfaceMeta,
		Spec: api.InterfaceSpec{
			VNI:           iface.Spec.VNI,
			Device:        iface.Spec.Device,
			IPs:           iface.Spec.IPs,
			UnderlayRoute: &underlayRoute,
			VirtualFunction: &api.VirtualFunction{
				Name:     res.Vf.Name,
				Domain:   res.Vf.Domain,
				Bus:      res.Vf.Bus,
				Slot:     res.Vf.Slot,
				Function: res.Vf.Function,
			},
			PXE: &api.PXE{Server: iface.Spec.PXE.Server, FileName: iface.Spec.PXE.FileName},
		},
		Status: api.ProtoStatusToStatus(res.Response.Status),
	}, nil
}

func (c *client) DeleteInterface(ctx context.Context, id string) (*api.Interface, error) {
	res, err := c.DPDKonmetalClient.DeleteInterface(ctx, &dpdkproto.InterfaceIDMsg{InterfaceID: []byte(id)})
	if err != nil {
		return &api.Interface{}, err
	}
	retInterface := &api.Interface{
		TypeMeta:      api.TypeMeta{Kind: api.InterfaceKind},
		InterfaceMeta: api.InterfaceMeta{ID: id},
		Status:        api.ProtoStatusToStatus(res),
	}
	if errorCode := res.GetError(); errorCode != 0 {
		return retInterface, errors.ErrServerError
	}
	return retInterface, nil
}

func (c *client) GetVirtualIP(ctx context.Context, interfaceID string) (*api.VirtualIP, error) {
	res, err := c.DPDKonmetalClient.GetInterfaceVIP(ctx, &dpdkproto.InterfaceIDMsg{
		InterfaceID: []byte(interfaceID),
	})
	if err != nil {
		return &api.VirtualIP{}, err
	}
	if errorCode := res.GetStatus().GetError(); errorCode != 0 {
		return &api.VirtualIP{
			TypeMeta:      api.TypeMeta{Kind: api.VirtualIPKind},
			VirtualIPMeta: api.VirtualIPMeta{InterfaceID: interfaceID},
			Status:        api.ProtoStatusToStatus(res.Status)}, errors.ErrServerError
	}
	return api.ProtoVirtualIPToVirtualIP(interfaceID, res)
}

func (c *client) AddVirtualIP(ctx context.Context, virtualIP *api.VirtualIP) (*api.VirtualIP, error) {
	res, err := c.DPDKonmetalClient.AddInterfaceVIP(ctx, &dpdkproto.InterfaceVIPMsg{
		InterfaceID: []byte(virtualIP.InterfaceID),
		InterfaceVIPIP: &dpdkproto.InterfaceVIPIP{
			IpVersion: api.NetIPAddrToProtoIPVersion(virtualIP.Spec.IP),
			Address:   []byte(virtualIP.Spec.IP.String()),
		},
	})
	if err != nil {
		return &api.VirtualIP{}, err
	}
	retVirtualIP := &api.VirtualIP{
		TypeMeta:      api.TypeMeta{Kind: api.VirtualIPKind},
		VirtualIPMeta: virtualIP.VirtualIPMeta,
		Spec: api.VirtualIPSpec{
			IP: virtualIP.Spec.IP,
		},
		Status: api.ProtoStatusToStatus(res.Status),
	}
	if errorCode := res.GetStatus().GetError(); errorCode != 0 {
		return retVirtualIP, errors.ErrServerError
	}
	underlayRoute, err := netip.ParseAddr(string(res.GetUnderlayRoute()))
	if err != nil {
		return retVirtualIP, fmt.Errorf("error parsing underlay route: %w", err)
	}
	retVirtualIP.Spec.UnderlayRoute = &underlayRoute
	return retVirtualIP, nil
}

func (c *client) DeleteVirtualIP(ctx context.Context, interfaceID string) (*api.VirtualIP, error) {
	res, err := c.DPDKonmetalClient.DeleteInterfaceVIP(ctx, &dpdkproto.InterfaceIDMsg{
		InterfaceID: []byte(interfaceID),
	})
	if err != nil {
		return &api.VirtualIP{}, err
	}
	retVirtualIP := &api.VirtualIP{
		TypeMeta:      api.TypeMeta{Kind: api.VirtualIPKind},
		VirtualIPMeta: api.VirtualIPMeta{InterfaceID: interfaceID},
	}
	if errorCode := res.GetError(); errorCode != 0 {
		return retVirtualIP, errors.ErrServerError
	}
	return retVirtualIP, nil
}

func (c *client) ListPrefixes(ctx context.Context, interfaceID string) (*api.PrefixList, error) {
	res, err := c.DPDKonmetalClient.ListInterfacePrefixes(ctx, &dpdkproto.InterfaceIDMsg{
		InterfaceID: []byte(interfaceID),
	})
	if err != nil {
		return nil, err
	}

	prefixes := make([]api.Prefix, len(res.GetPrefixes()))
	for i, dpdkPrefix := range res.GetPrefixes() {
		prefix, err := api.ProtoPrefixToPrefix(interfaceID, dpdkPrefix)
		if err != nil {
			return nil, err
		}

		prefixes[i] = *prefix
	}

	return &api.PrefixList{
		TypeMeta:       api.TypeMeta{Kind: api.PrefixListKind},
		PrefixListMeta: api.PrefixListMeta{InterfaceID: interfaceID},
		Items:          prefixes,
	}, nil
}

func (c *client) AddPrefix(ctx context.Context, prefix *api.Prefix) (*api.Prefix, error) {
	res, err := c.DPDKonmetalClient.AddInterfacePrefix(ctx, &dpdkproto.InterfacePrefixMsg{
		InterfaceID: &dpdkproto.InterfaceIDMsg{
			InterfaceID: []byte(prefix.InterfaceID),
		},
		Prefix: &dpdkproto.Prefix{
			IpVersion:    api.NetIPAddrToProtoIPVersion(prefix.Spec.Prefix.Addr()),
			Address:      []byte(prefix.Spec.Prefix.Addr().String()),
			PrefixLength: uint32(prefix.Spec.Prefix.Bits()),
		},
	})
	if err != nil {
		return &api.Prefix{}, err
	}
	retPrefix := &api.Prefix{
		TypeMeta:   api.TypeMeta{Kind: api.PrefixKind},
		PrefixMeta: prefix.PrefixMeta,
		Spec:       api.PrefixSpec{Prefix: prefix.Spec.Prefix},
		Status:     api.ProtoStatusToStatus(res.Status),
	}

	if errorCode := res.GetStatus().GetError(); errorCode != 0 {
		return retPrefix, errors.ErrServerError
	}
	underlayRoute, err := netip.ParseAddr(string(res.GetUnderlayRoute()))
	if err != nil {
		return retPrefix, fmt.Errorf("error parsing underlay route: %w", err)
	}
	retPrefix.Spec.UnderlayRoute = &underlayRoute
	return retPrefix, nil
}

func (c *client) DeletePrefix(ctx context.Context, interfaceID string, prefix netip.Prefix) (*api.Prefix, error) {
	res, err := c.DPDKonmetalClient.DeleteInterfacePrefix(ctx, &dpdkproto.InterfacePrefixMsg{
		InterfaceID: &dpdkproto.InterfaceIDMsg{
			InterfaceID: []byte(interfaceID),
		},
		Prefix: &dpdkproto.Prefix{
			IpVersion:    api.NetIPAddrToProtoIPVersion(prefix.Addr()),
			Address:      []byte(prefix.Addr().String()),
			PrefixLength: uint32(prefix.Bits()),
		},
	})
	if err != nil {
		return &api.Prefix{}, err
	}
	retPrefix := &api.Prefix{
		TypeMeta:   api.TypeMeta{Kind: api.PrefixKind},
		PrefixMeta: api.PrefixMeta{InterfaceID: interfaceID},
		Spec:       api.PrefixSpec{Prefix: prefix},
		Status:     api.ProtoStatusToStatus(res),
	}
	if errorCode := res.GetError(); errorCode != 0 {
		return retPrefix, errors.ErrServerError
	}
	return retPrefix, nil
}

func (c *client) AddRoute(ctx context.Context, route *api.Route) (*api.Route, error) {
	res, err := c.DPDKonmetalClient.AddRoute(ctx, &dpdkproto.VNIRouteMsg{
		Vni: &dpdkproto.VNIMsg{Vni: route.VNI},
		Route: &dpdkproto.Route{
			IpVersion: api.NetIPAddrToProtoIPVersion(*route.Spec.NextHop.IP),
			Weight:    100,
			Prefix: &dpdkproto.Prefix{
				IpVersion:    api.NetIPAddrToProtoIPVersion(route.Spec.Prefix.Addr()),
				Address:      []byte(route.Spec.Prefix.Addr().String()),
				PrefixLength: uint32(route.Spec.Prefix.Bits()),
			},
			NexthopVNI:     route.Spec.NextHop.VNI,
			NexthopAddress: []byte(route.Spec.NextHop.IP.String()),
		},
	})
	if err != nil {
		return &api.Route{}, err
	}
	retRoute := &api.Route{
		TypeMeta:  api.TypeMeta{Kind: api.RouteKind},
		RouteMeta: route.RouteMeta,
		Spec: api.RouteSpec{
			Prefix:  route.Spec.Prefix,
			NextHop: &api.RouteNextHop{}},
		Status: api.ProtoStatusToStatus(res),
	}
	if errorCode := res.GetError(); errorCode != 0 {
		return retRoute, errors.ErrServerError
	}
	retRoute.Spec = route.Spec
	return retRoute, nil
}

func (c *client) DeleteRoute(ctx context.Context, vni uint32, prefix netip.Prefix) (*api.Route, error) {
	res, err := c.DPDKonmetalClient.DeleteRoute(ctx, &dpdkproto.VNIRouteMsg{
		Vni: &dpdkproto.VNIMsg{Vni: vni},
		Route: &dpdkproto.Route{
			IpVersion: api.NetIPAddrToProtoIPVersion(prefix.Addr()),
			Weight:    100,
			Prefix: &dpdkproto.Prefix{
				IpVersion:    api.NetIPAddrToProtoIPVersion(prefix.Addr()),
				Address:      []byte(prefix.Addr().String()),
				PrefixLength: uint32(prefix.Bits()),
			},
		},
	})
	if err != nil {
		return &api.Route{}, err
	}
	retRoute := &api.Route{
		TypeMeta:  api.TypeMeta{Kind: api.RouteKind},
		RouteMeta: api.RouteMeta{VNI: vni},
		Spec: api.RouteSpec{
			Prefix:  &prefix,
			NextHop: &api.RouteNextHop{},
		},
		Status: api.ProtoStatusToStatus(res),
	}
	if errorCode := res.GetError(); errorCode != 0 {
		return retRoute, errors.ErrServerError
	}
	return retRoute, nil
}

func (c *client) ListRoutes(ctx context.Context, vni uint32) (*api.RouteList, error) {
	res, err := c.DPDKonmetalClient.ListRoutes(ctx, &dpdkproto.VNIMsg{
		Vni: vni,
	})
	if err != nil {
		return nil, err
	}

	routes := make([]api.Route, len(res.GetRoutes()))
	for i, dpdkRoute := range res.GetRoutes() {
		route, err := api.ProtoRouteToRoute(vni, dpdkRoute)
		if err != nil {
			return nil, err
		}

		routes[i] = *route
	}

	return &api.RouteList{
		TypeMeta:      api.TypeMeta{Kind: api.RouteListKind},
		RouteListMeta: api.RouteListMeta{VNI: vni},
		Items:         routes,
	}, nil
}

func (c *client) GetNat(ctx context.Context, interfaceID string) (*api.Nat, error) {
	res, err := c.DPDKonmetalClient.GetNAT(ctx, &dpdkproto.GetNATRequest{InterfaceID: []byte(interfaceID)})
	if err != nil {
		return &api.Nat{}, err
	}
	if errorCode := res.GetStatus().GetError(); errorCode != 0 {
		return &api.Nat{
			TypeMeta: api.TypeMeta{Kind: api.NatKind},
			NatMeta:  api.NatMeta{InterfaceID: interfaceID},
			Status:   api.ProtoStatusToStatus(res.Status)}, errors.ErrServerError
	}
	return api.ProtoNatToNat(res, interfaceID)
}

func (c *client) AddNat(ctx context.Context, nat *api.Nat) (*api.Nat, error) {
	res, err := c.DPDKonmetalClient.AddNAT(ctx, &dpdkproto.AddNATRequest{
		InterfaceID: []byte(nat.NatMeta.InterfaceID),
		NatVIPIP: &dpdkproto.NATIP{
			IpVersion: api.NetIPAddrToProtoIPVersion(*nat.Spec.NatVIPIP),
			Address:   []byte(nat.Spec.NatVIPIP.String()),
		},
		MinPort: nat.Spec.MinPort,
		MaxPort: nat.Spec.MaxPort,
	})
	if err != nil {
		return &api.Nat{}, err
	}
	retNat := &api.Nat{
		TypeMeta: api.TypeMeta{Kind: api.NatKind},
		NatMeta:  nat.NatMeta,
		Status:   api.ProtoStatusToStatus(res.Status),
	}
	if errorCode := res.GetStatus().GetError(); errorCode != 0 {
		return retNat, errors.ErrServerError
	}

	underlayRoute, err := netip.ParseAddr(string(res.GetUnderlayRoute()))
	if err != nil {
		return retNat, fmt.Errorf("error parsing underlay route: %w", err)
	}

	retNat.Spec = nat.Spec
	retNat.Spec.UnderlayRoute = &underlayRoute
	return retNat, nil
}

func (c *client) DeleteNat(ctx context.Context, interfaceID string) (*api.Nat, error) {
	res, err := c.DPDKonmetalClient.DeleteNAT(ctx, &dpdkproto.DeleteNATRequest{
		InterfaceID: []byte(interfaceID),
	})
	if err != nil {
		return &api.Nat{}, err
	}
	retNat := &api.Nat{
		TypeMeta: api.TypeMeta{Kind: api.NatKind},
		NatMeta:  api.NatMeta{InterfaceID: interfaceID},
		Status:   api.ProtoStatusToStatus(res),
	}
	if errorCode := res.GetError(); errorCode != 0 {
		return retNat, errors.ErrServerError
	}
	return retNat, nil
}

func (c *client) AddNeighborNat(ctx context.Context, nNat *api.NeighborNat) (*api.NeighborNat, error) {

	res, err := c.DPDKonmetalClient.AddNeighborNAT(ctx, &dpdkproto.AddNeighborNATRequest{
		NatVIPIP: &dpdkproto.NATIP{
			IpVersion: api.NetIPAddrToProtoIPVersion(*nNat.NeighborNatMeta.NatVIPIP),
			Address:   []byte(nNat.NeighborNatMeta.NatVIPIP.String()),
		},
		Vni:           nNat.Spec.Vni,
		MinPort:       nNat.Spec.MinPort,
		MaxPort:       nNat.Spec.MaxPort,
		UnderlayRoute: []byte(nNat.Spec.UnderlayRoute.String()),
	})
	if err != nil {
		return &api.NeighborNat{}, err
	}
	retnNat := &api.NeighborNat{
		TypeMeta:        api.TypeMeta{Kind: api.NeighborNatKind},
		NeighborNatMeta: nNat.NeighborNatMeta,
		Status:          api.ProtoStatusToStatus(res),
	}
	if errorCode := res.GetError(); errorCode != 0 {
		return retnNat, errors.ErrServerError
	}
	retnNat.Spec = nNat.Spec
	return retnNat, nil
}

func (c *client) GetNATInfo(ctx context.Context, natVIPIP netip.Addr, natType string) (*api.NatList, error) {
	var nType int32
	switch strings.ToLower(natType) {
	case "local", "1":
		nType = 1
	case "neigh", "2", "neighbor":
		nType = 2
	case "any", "0", "":
		nType = 0
	default:
		return nil, fmt.Errorf("nat info type can be only: Any = 0/Local = 1/Neigh(bor) = 2")
	}

	req := dpdkproto.GetNATInfoRequest{
		NatVIPIP: &dpdkproto.NATIP{IpVersion: api.NetIPAddrToProtoIPVersion(natVIPIP),
			Address: []byte(natVIPIP.String()),
		},
		NatInfoType: dpdkproto.NATInfoType(nType),
	}
	// nat info type not defined, try both types
	res := &dpdkproto.GetNATInfoResponse{NatVIPIP: &dpdkproto.NATIP{}}
	var err error
	if nType == 0 {
		req.NatInfoType = 1
		res1, err1 := c.DPDKonmetalClient.GetNATInfo(ctx, &req)
		if err1 != nil {
			return nil, err1
		}
		req.NatInfoType = 2
		res2, err2 := c.DPDKonmetalClient.GetNATInfo(ctx, &req)
		if err2 != nil {
			return nil, err2
		}
		res.NatInfoEntries = append(res.NatInfoEntries, res1.NatInfoEntries...)
		res.NatInfoEntries = append(res.NatInfoEntries, res2.NatInfoEntries...)
		res.NatVIPIP.Address = []byte(natVIPIP.String())
	} else {
		res, err = c.DPDKonmetalClient.GetNATInfo(ctx, &req)
		if err != nil {
			return nil, err
		}
	}

	var nats = make([]api.Nat, len(res.NatInfoEntries))
	var nat api.Nat
	for i, natInfoEntry := range res.GetNatInfoEntries() {

		var underlayRoute, vipIP netip.Addr
		if natInfoEntry.UnderlayRoute != nil {
			underlayRoute, err = netip.ParseAddr(string(natInfoEntry.GetUnderlayRoute()))
			if err != nil {
				return nil, fmt.Errorf("error parsing underlay route: %w", err)
			}
			nat.Spec.UnderlayRoute = &underlayRoute
			vipIP, err = netip.ParseAddr(string(res.NatVIPIP.Address))
			if err != nil {
				return nil, fmt.Errorf("error parsing vip ip: %w", err)
			}
			nat.Spec.NatVIPIP = &vipIP
		} else {
			nat.Spec.NatVIPIP = nil
		}
		nat.Kind = api.NatKind
		nat.Spec.MinPort = natInfoEntry.MinPort
		nat.Spec.MaxPort = natInfoEntry.MaxPort
		nats[i] = nat
	}
	return &api.NatList{
		TypeMeta:    api.TypeMeta{Kind: api.NatListKind},
		NatListMeta: api.NatListMeta{NatIP: &natVIPIP, NatInfoType: natType},
		Items:       nats,
	}, nil
}

func (c *client) DeleteNeighborNat(ctx context.Context, neigbhorNat api.NeighborNat) (*api.NeighborNat, error) {
	res, err := c.DPDKonmetalClient.DeleteNeighborNAT(ctx, &dpdkproto.DeleteNeighborNATRequest{
		NatVIPIP: &dpdkproto.NATIP{
			IpVersion: api.NetIPAddrToProtoIPVersion(*neigbhorNat.NatVIPIP),
			Address:   []byte(neigbhorNat.NatVIPIP.String()),
		},
		Vni:     neigbhorNat.Spec.Vni,
		MinPort: neigbhorNat.Spec.MinPort,
		MaxPort: neigbhorNat.Spec.MaxPort,
	})
	if err != nil {
		return &api.NeighborNat{}, err
	}
	nnat := &api.NeighborNat{
		TypeMeta:        api.TypeMeta{Kind: api.NeighborNatKind},
		NeighborNatMeta: neigbhorNat.NeighborNatMeta,
		Status:          api.ProtoStatusToStatus(res),
	}
	if errorCode := res.GetError(); errorCode != 0 {
		return nnat, errors.ErrServerError
	}
	return nnat, nil
}

func (c *client) ListFirewallRules(ctx context.Context, interfaceID string) (*api.FirewallRuleList, error) {
	res, err := c.DPDKonmetalClient.ListFirewallRules(ctx, &dpdkproto.ListFirewallRulesRequest{
		InterfaceID: []byte(interfaceID),
	})
	if err != nil {
		return &api.FirewallRuleList{}, err
	}

	fwRules := make([]api.FirewallRule, len(res.GetRules()))
	for i, dpdkFwRule := range res.GetRules() {
		fwRule, err := api.ProtoFwRuleToFwRule(dpdkFwRule, interfaceID)
		if err != nil {
			return &api.FirewallRuleList{}, err
		}
		fwRules[i] = *fwRule
	}

	return &api.FirewallRuleList{
		TypeMeta:             api.TypeMeta{Kind: api.FirewallRuleListKind},
		FirewallRuleListMeta: api.FirewallRuleListMeta{InterfaceID: interfaceID},
		Items:                fwRules,
	}, nil
}

func (c *client) AddFirewallRule(ctx context.Context, fwRule *api.FirewallRule) (*api.FirewallRule, error) {
	var action, direction, ipv uint8

	switch strings.ToLower(fwRule.Spec.FirewallAction) {
	case "accept", "allow", "1":
		action = 1
		fwRule.Spec.FirewallAction = "Accept"
	case "drop", "deny", "0":
		action = 0
		fwRule.Spec.FirewallAction = "Drop"
	default:
		return &api.FirewallRule{}, fmt.Errorf("firewall action can be only: drop/deny/0|accept/allow/1")
	}

	switch strings.ToLower(fwRule.Spec.TrafficDirection) {
	case "ingress", "0":
		direction = 0
		fwRule.Spec.TrafficDirection = "Ingress"
	case "egress", "1":
		direction = 1
		fwRule.Spec.TrafficDirection = "Egress"
	default:
		return &api.FirewallRule{}, fmt.Errorf("traffic direction can be only: Ingress = 0/Egress = 1")
	}

	switch strings.ToLower(fwRule.Spec.IpVersion) {
	case "ipv4", "0":
		ipv = 0
		fwRule.Spec.IpVersion = "IPv4"
	case "ipv6", "1":
		ipv = 1
		fwRule.Spec.IpVersion = "IPv6"
	default:
		return &api.FirewallRule{}, fmt.Errorf("ip version can be only: IPv4 = 0/IPv6 = 1")
	}

	req := dpdkproto.AddFirewallRuleRequest{
		InterfaceID: []byte(fwRule.FirewallRuleMeta.InterfaceID),
		Rule: &dpdkproto.FirewallRule{
			RuleID:    []byte(fwRule.Spec.RuleID),
			Direction: dpdkproto.TrafficDirection(direction),
			Action:    dpdkproto.FirewallAction(action),
			Priority:  fwRule.Spec.Priority,
			IpVersion: dpdkproto.IPVersion(ipv),
			SourcePrefix: &dpdkproto.Prefix{
				IpVersion:    dpdkproto.IPVersion(ipv),
				Address:      []byte(fwRule.Spec.SourcePrefix.Addr().String()),
				PrefixLength: uint32(fwRule.Spec.SourcePrefix.Bits()),
			},
			DestinationPrefix: &dpdkproto.Prefix{
				IpVersion:    dpdkproto.IPVersion(ipv),
				Address:      []byte(fwRule.Spec.DestinationPrefix.Addr().String()),
				PrefixLength: uint32(fwRule.Spec.DestinationPrefix.Bits()),
			},
			ProtocolFilter: fwRule.Spec.ProtocolFilter,
		},
	}

	res, err := c.DPDKonmetalClient.AddFirewallRule(ctx, &req)
	if err != nil {
		return &api.FirewallRule{}, err
	}
	retFwrule := &api.FirewallRule{
		TypeMeta:         api.TypeMeta{Kind: api.FirewallRuleKind},
		FirewallRuleMeta: api.FirewallRuleMeta{InterfaceID: fwRule.InterfaceID},
		Spec:             api.FirewallRuleSpec{RuleID: fwRule.Spec.RuleID},
		Status:           api.ProtoStatusToStatus(res.Status)}
	if res.Status.Error != 0 {
		return retFwrule, errors.ErrServerError
	}
	retFwrule.Spec = fwRule.Spec
	return retFwrule, nil
}

func (c *client) GetFirewallRule(ctx context.Context, ruleID string, interfaceID string) (*api.FirewallRule, error) {
	res, err := c.DPDKonmetalClient.GetFirewallRule(ctx, &dpdkproto.GetFirewallRuleRequest{
		InterfaceID: []byte(interfaceID),
		RuleID:      []byte(ruleID),
	})
	if err != nil {
		return &api.FirewallRule{}, err
	}
	if errorCode := res.GetStatus().GetError(); errorCode != 0 {
		return &api.FirewallRule{
			TypeMeta:         api.TypeMeta{Kind: api.FirewallRuleKind},
			FirewallRuleMeta: api.FirewallRuleMeta{InterfaceID: interfaceID},
			Spec:             api.FirewallRuleSpec{RuleID: ruleID},
			Status:           api.ProtoStatusToStatus(res.Status)}, errors.ErrServerError
	}

	return api.ProtoFwRuleToFwRule(res.Rule, interfaceID)
}

func (c *client) DeleteFirewallRule(ctx context.Context, interfaceID string, ruleID string) (*api.FirewallRule, error) {
	res, err := c.DPDKonmetalClient.DeleteFirewallRule(ctx, &dpdkproto.DeleteFirewallRuleRequest{
		InterfaceID: []byte(interfaceID),
		RuleID:      []byte(ruleID),
	})
	if err != nil {
		return &api.FirewallRule{}, err
	}
	retFwrule := &api.FirewallRule{
		TypeMeta:         api.TypeMeta{Kind: api.FirewallRuleKind},
		FirewallRuleMeta: api.FirewallRuleMeta{InterfaceID: interfaceID},
		Spec:             api.FirewallRuleSpec{RuleID: ruleID},
		Status:           api.ProtoStatusToStatus(res),
	}
	if errorCode := res.GetError(); errorCode != 0 {
		return retFwrule, errors.ErrServerError
	}
	return retFwrule, nil
}

func (c *client) Initialized(ctx context.Context) (string, error) {
	res, err := c.DPDKonmetalClient.Initialized(ctx, &dpdkproto.Empty{})
	if err != nil {
		return "", err
	}
	return res.Uuid, nil
}

func (c *client) Init(ctx context.Context, initConfig dpdkproto.InitConfig) (*api.Init, error) {
	res, err := c.DPDKonmetalClient.Init(ctx, &initConfig)
	if err != nil {
		return &api.Init{}, err
	}
	if errorCode := res.GetError(); errorCode != 0 {
		return &api.Init{Status: api.ProtoStatusToStatus(res)}, errors.ErrServerError
	}
	return &api.Init{TypeMeta: api.TypeMeta{Kind: "Init"}, Status: api.ProtoStatusToStatus(res)}, nil
}

func (c *client) GetVni(ctx context.Context, vni uint32, vniType uint8) (*api.Vni, error) {
	res, err := c.DPDKonmetalClient.IsVniInUse(ctx, &dpdkproto.IsVniInUseRequest{
		Vni:  vni,
		Type: dpdkproto.VniType(vniType),
	})
	if err != nil {
		return &api.Vni{}, err
	}
	retVni := &api.Vni{
		TypeMeta: api.TypeMeta{Kind: api.VniKind},
		VniMeta:  api.VniMeta{VNI: vni, VniType: vniType},
		Status:   api.ProtoStatusToStatus(res.Status),
	}
	if errorCode := res.GetStatus().GetError(); errorCode != 0 {
		return retVni, errors.ErrServerError
	}
	retVni.Spec.InUse = res.InUse
	return retVni, nil
}

func (c *client) ResetVni(ctx context.Context, vni uint32, vniType uint8) (*api.Vni, error) {
	res, err := c.DPDKonmetalClient.ResetVni(ctx, &dpdkproto.ResetVniRequest{
		Vni:  vni,
		Type: dpdkproto.VniType(vniType),
	})
	if err != nil {
		return &api.Vni{}, err
	}
	retVni := &api.Vni{
		TypeMeta: api.TypeMeta{Kind: api.VniKind},
		VniMeta:  api.VniMeta{VNI: vni, VniType: vniType},
		Status:   api.ProtoStatusToStatus(res),
	}
	if errorCode := res.GetError(); errorCode != 0 {
		return retVni, errors.ErrServerError
	}
	return retVni, nil
}
