package service

import (
	"fmt"
	"net/netip"

	"github.com/docker/go-connections/nat"
	"github.com/moby/moby/api/types/network"
)

func networkPort(p nat.Port) (network.Port, error) {
	return network.ParsePort(string(p))
}

func networkPortMap(pm nat.PortMap) (network.PortMap, error) {
	if len(pm) == 0 {
		return nil, nil
	}
	result := make(network.PortMap, len(pm))
	for port, bindings := range pm {
		np, err := networkPort(port)
		if err != nil {
			return nil, fmt.Errorf("convert port %q: %w", port, err)
		}
		converted := make([]network.PortBinding, len(bindings))
		for i, binding := range bindings {
			hostIP := binding.HostIP
			addr, err := netip.ParseAddr(hostIP)
			if err != nil {
				if hostIP == "" {
					addr = netip.IPv4Unspecified()
				} else {
					return nil, fmt.Errorf("parse host IP %q: %w", hostIP, err)
				}
			}
			converted[i] = network.PortBinding{
				HostIP:   addr,
				HostPort: binding.HostPort,
			}
		}
		result[np] = converted
	}
	return result, nil
}

func networkPortSet(ports map[nat.Port]struct{}) (network.PortSet, error) {
	if len(ports) == 0 {
		return nil, nil
	}
	result := make(network.PortSet, len(ports))
	for port := range ports {
		np, err := networkPort(port)
		if err != nil {
			return nil, fmt.Errorf("convert port %q: %w", port, err)
		}
		result[np] = struct{}{}
	}
	return result, nil
}

func networkDNSAddrs(servers []string) ([]netip.Addr, error) {
	if len(servers) == 0 {
		return nil, nil
	}
	result := make([]netip.Addr, 0, len(servers))
	for _, server := range servers {
		addr, err := netip.ParseAddr(server)
		if err != nil {
			return nil, fmt.Errorf("parse DNS server %q: %w", server, err)
		}
		result = append(result, addr)
	}
	return result, nil
}

func endpointIP(settings *network.EndpointSettings) string {
	if settings == nil || !settings.IPAddress.IsValid() {
		return ""
	}
	return settings.IPAddress.String()
}
