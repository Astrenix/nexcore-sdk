package nexcore

import (
	"encoding/json"
)

// AccountNamespace implements the v1 账户 read endpoints.
//
// 对应 NexCore 账户模块的公开只读接口:
//
//	GET /api/v1/account/balance          GetBalance        查询账户余额
//	GET /api/v1/account/deposit-address  GetDepositAddress 查询/获取充值地址
//
// 鉴权:X-API-Key + X-Secret-Key 双 header(用 cfg.APIKey / cfg.APISecret).
// account 与 vcard 命名空间共用同一对 MPK 商户密钥.
type AccountNamespace struct {
	c *Client
}

// headers 构造 X-API-Key + X-Secret-Key(复用与 energy 一致的双密钥读模式).
func (n *AccountNamespace) headers() (map[string]string, error) {
	if n.c.cfg.APIKey == "" || n.c.cfg.APISecret == "" {
		return nil, &Error{Message: "APIKey / APISecret not configured", Code: -1}
	}
	return map[string]string{
		"X-API-Key":    n.c.cfg.APIKey,
		"X-Secret-Key": n.c.cfg.APISecret,
	}, nil
}

// GetBalance returns the account balance.
//
// GET /api/v1/account/balance
func (n *AccountNamespace) GetBalance() (json.RawMessage, error) {
	h, err := n.headers()
	if err != nil {
		return nil, err
	}
	return n.c.transport.do("GET", "/api/v1/account/balance", &requestOpts{Headers: h})
}

// GetDepositAddress returns the account deposit address(es).
//
// GET /api/v1/account/deposit-address
func (n *AccountNamespace) GetDepositAddress() (json.RawMessage, error) {
	h, err := n.headers()
	if err != nil {
		return nil, err
	}
	return n.c.transport.do("GET", "/api/v1/account/deposit-address", &requestOpts{Headers: h})
}
