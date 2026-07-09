package service

import (
	"net/netip"
	"testing"

	ctr "github.com/moby/moby/api/types/container"
	"github.com/moby/moby/api/types/network"
	"github.com/docker/go-connections/nat"
	"github.com/stretchr/testify/require"
)

func TestHostPortAddressInDockerCustomNetworkWithoutBindings(t *testing.T) {
	t.Parallel()

	port, err := nat.NewPort("tcp", "4444")
	require.NoError(t, err)

	seleniumPort, err := network.ParsePort("4444/tcp")
	require.NoError(t, err)

	stat := ctr.InspectResponse{
		NetworkSettings: &ctr.NetworkSettings{
			Ports: network.PortMap{
				seleniumPort: nil,
			},
			Networks: map[string]*network.EndpointSettings{
				"selenoid": {
					IPAddress: netip.MustParseAddr("172.18.0.3"),
				},
			},
		},
	}

	addr := hostPortAddress(Environment{
		InDocker: true,
		Network:  "selenoid",
	}, stat, port)
	require.Equal(t, "172.18.0.3:4444", addr)
}
