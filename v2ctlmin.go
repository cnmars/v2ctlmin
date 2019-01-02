package v2ctlmin

import (
	"context"
	"fmt"
	"log"

	"google.golang.org/grpc"
	proxymancmd "v2ray.com/core/app/proxyman/command"
	statscmd "v2ray.com/core/app/stats/command"
	"v2ray.com/core/common/protocol"
	"v2ray.com/core/common/serial"
	"v2ray.com/core/common/uuid"
	"v2ray.com/core/proxy/vmess"
)

func GenerateUUID() string {
	u := uuid.New()
	return u.String()
}

type ServiceClient struct {
	APIAddress  string
	APIPort     uint32
	statClient  statscmd.StatsServiceClient
	proxyClient proxymancmd.HandlerServiceClient
}

// NewServiceClient ...
// Generate Stats service client obj
func NewServiceClient(addr string, port uint32) *ServiceClient {
	cmdConn, err := grpc.Dial(fmt.Sprintf("%s:%d", addr, port), grpc.WithInsecure())
	if err != nil {
		log.Fatal(err)
		return nil
	}

	svr := ServiceClient{APIAddress: addr, APIPort: port,
		statClient:  statscmd.NewStatsServiceClient(cmdConn),
		proxyClient: proxymancmd.NewHandlerServiceClient(cmdConn),
	}
	return &svr
}

func (h *ServiceClient) QueryStats(pattern string, reset bool) map[string]int64 {
	sresp, err := h.statClient.QueryStats(context.Background(), &statscmd.QueryStatsRequest{
		Pattern: pattern,
		Reset_:  reset,
	})

	result := make(map[string]int64)
	if err != nil {
		log.Printf("failed to call grpc command: %v", err)
	} else {
		// log.Printf("%v", sresp)
		for _, stat := range sresp.Stat {
			result[stat.Name] = stat.Value
		}
	}

	return result
}

func (h *ServiceClient) GetStats(name string, reset bool) (string, int64) {
	sresp, err := h.statClient.GetStats(context.Background(), &statscmd.GetStatsRequest{
		Name:   name,
		Reset_: reset,
	})

	if err != nil {
		log.Printf("%v", err)
		return "", 0
	}

	return sresp.Stat.Name, sresp.Stat.Value
}

func (h *ServiceClient) AddUser(inboundTag string, email string, level uint32, uuid string, alterID uint32) {

	resp, err := h.proxyClient.AlterInbound(context.Background(), &proxymancmd.AlterInboundRequest{
		Tag: inboundTag,
		Operation: serial.ToTypedMessage(&proxymancmd.AddUserOperation{
			User: &protocol.User{
				Level: level,
				Email: email,
				Account: serial.ToTypedMessage(&vmess.Account{
					Id:               uuid,
					AlterId:          alterID,
					SecuritySettings: &protocol.SecurityConfig{Type: protocol.SecurityType_AUTO},
				}),
			},
		}),
	})

	if err != nil {
		log.Printf("%v", err)
	} else {
		log.Printf("%v", resp)
	}
}

func (h *ServiceClient) RemoveUser(inboundTag string, email string) {
	resp, err := h.proxyClient.AlterInbound(context.Background(), &proxymancmd.AlterInboundRequest{
		Tag: inboundTag,
		Operation: serial.ToTypedMessage(&proxymancmd.RemoveUserOperation{
			Email: email,
		}),
	})

	if err != nil {
		log.Printf("%v", err)
	} else {
		log.Printf("%v", resp)
	}
}
