package nexcore

import (
	"encoding/json"
	"fmt"
)

// EnergyNamespace implements all 8 TRON energy lease v1 endpoints.
//
// 对应 /docs 文档 "能量租赁" 模块的全部 v1 公开接口
// (对照 internal/handler/trxx_api.go):
//
//	GET  /api/v1/energy/info             GetInfo             平台公开信息
//	GET  /api/v1/energy/price            GetPrice            指定能量+周期的报价
//	GET  /api/v1/energy/estimate-energy  EstimateEnergy      根据地址估算 TRC20 转账所需能量
//	POST /api/v1/energy/order            CreateOrder         创建常规租赁订单
//	POST /api/v1/energy/order/onetime    CreateOnetimeOrder  创建一次性订单
//	GET  /api/v1/energy/order/:serial    QueryOrder          查询订单(serial 字符串)
//	GET  /api/v1/energy/orders           ListOrders          列出所有订单
//	POST /api/v1/energy/order/reclaim    ReclaimOrder        主动回收订单
//
// 鉴权:X-API-Key + X-Secret-Key 双 header.
type EnergyNamespace struct {
	c *Client
}

// headers 构造 X-API-Key + X-Secret-Key.
func (n *EnergyNamespace) headers() (map[string]string, error) {
	if n.c.cfg.EnergyAPIKey == "" || n.c.cfg.EnergySecretKey == "" {
		return nil, &Error{Message: "EnergyAPIKey / EnergySecretKey not configured", Code: -1}
	}
	return map[string]string{
		"X-API-Key":    n.c.cfg.EnergyAPIKey,
		"X-Secret-Key": n.c.cfg.EnergySecretKey,
	}, nil
}

// GetInfo returns platform public info (available energy, pricing tiers, etc).
//
// GET /api/v1/energy/info
//
// 返回 {platform_avail_energy, minimum_order_energy, maximum_order_energy, tiered_pricing, ...}.
func (n *EnergyNamespace) GetInfo() (json.RawMessage, error) {
	h, err := n.headers()
	if err != nil {
		return nil, err
	}
	return n.c.transport.do("GET", "/api/v1/energy/info", &requestOpts{Headers: h})
}

// GetPrice quotes the price for a given energy amount and period.
//
// GET /api/v1/energy/price?energy=65000&period=1D
//
// period: "1H" / "6H" / "1D" / "3D" / "1W",空字符串视为 "1D".
func (n *EnergyNamespace) GetPrice(energy int, period string) (json.RawMessage, error) {
	if period == "" {
		period = "1D"
	}
	h, err := n.headers()
	if err != nil {
		return nil, err
	}
	return n.c.transport.do("GET", "/api/v1/energy/price", &requestOpts{
		Query:   map[string]any{"energy": energy, "period": period},
		Headers: h,
	})
}

// EstimateEnergy estimates the energy needed for a TRC20 transfer to the given address.
//
// GET /api/v1/energy/estimate-energy?receive_addr=TXxxxxxxxx
//
// 返回 {estimated_energy, has_usdt_balance, ...}.
func (n *EnergyNamespace) EstimateEnergy(receiveAddr string) (json.RawMessage, error) {
	h, err := n.headers()
	if err != nil {
		return nil, err
	}
	return n.c.transport.do("GET", "/api/v1/energy/estimate-energy", &requestOpts{
		Query:   map[string]any{"receive_addr": receiveAddr},
		Headers: h,
	})
}

// CreateOrder creates a regular lease order.
//
// POST /api/v1/energy/order
//
// 必填:
//   receive_addr (string) — 收能量的目标 TRON 地址
//   energy       (int)    — 能量数(>= minimum_order_energy)
//   period       (string) — 1H / 6H / 1D / 3D / 1W
//
// 可选:
//   out_serial   (string) — 商户侧订单号(幂等用)
//
// 返回 {serial, status, delegated_at, ...}.
func (n *EnergyNamespace) CreateOrder(params map[string]any) (json.RawMessage, error) {
	h, err := n.headers()
	if err != nil {
		return nil, err
	}
	return n.c.transport.do("POST", "/api/v1/energy/order", &requestOpts{Body: params, Headers: h})
}

// CreateOnetimeOrder creates a one-time order (energy is reclaimed after one use).
//
// POST /api/v1/energy/order/onetime
//
// 适用场景:用户只做一笔 TRC20 转账,转完即丢能量.
func (n *EnergyNamespace) CreateOnetimeOrder(params map[string]any) (json.RawMessage, error) {
	h, err := n.headers()
	if err != nil {
		return nil, err
	}
	return n.c.transport.do("POST", "/api/v1/energy/order/onetime", &requestOpts{Body: params, Headers: h})
}

// QueryOrder queries an order by serial string.
//
// GET /api/v1/energy/order/:serial
//
// 注意:serial 是字符串序列号,**不是**数字 id.
func (n *EnergyNamespace) QueryOrder(serial string) (json.RawMessage, error) {
	h, err := n.headers()
	if err != nil {
		return nil, err
	}
	return n.c.transport.do("GET", fmt.Sprintf("/api/v1/energy/order/%s", serial), &requestOpts{Headers: h})
}

// ListOrders lists all orders, optionally filtered.
//
// GET /api/v1/energy/orders
//
// filter 可包含 status / page / page_size 等.
func (n *EnergyNamespace) ListOrders(filter map[string]any) (json.RawMessage, error) {
	h, err := n.headers()
	if err != nil {
		return nil, err
	}
	return n.c.transport.do("GET", "/api/v1/energy/orders", &requestOpts{Query: filter, Headers: h})
}

// ReclaimOrder actively reclaims an order (returns energy to the platform).
//
// POST /api/v1/energy/order/reclaim
func (n *EnergyNamespace) ReclaimOrder(serial string) (json.RawMessage, error) {
	h, err := n.headers()
	if err != nil {
		return nil, err
	}
	return n.c.transport.do("POST", "/api/v1/energy/order/reclaim", &requestOpts{
		Body:    map[string]any{"serial": serial},
		Headers: h,
	})
}
