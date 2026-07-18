package vnpay

import (
	"crypto/hmac"
	"crypto/sha512"
	"encoding/hex"
	"fmt"
	"net/url"
	"sort"
	"strings"
)

type VNPayClient struct {
	tmnCode    string
	hashSecret string
	paymentURL string
	returnURL  string
}

func NewVNPayClient(tmnCode, hashSecret, paymentURL, returnURL string) *VNPayClient {
	if tmnCode == "" {
		tmnCode = "MOCK_TMN"
	}
	if hashSecret == "" {
		hashSecret = "MOCK_SECRET"
	}
	if paymentURL == "" {
		paymentURL = "https://sandbox.vnpayment.vn/paymentv2/vpcpay.html"
	}
	if returnURL == "" {
		returnURL = "http://localhost:3000/payment-result"
	}
	return &VNPayClient{
		tmnCode:    tmnCode,
		hashSecret: hashSecret,
		paymentURL: paymentURL,
		returnURL:  returnURL,
	}
}

// GeneratePaymentURL builds the VNPay checkout URL
func (c *VNPayClient) GeneratePaymentURL(txnRef string, amount int64, ipAddr, desc, createDate string) string {
	// VNPay requires amount * 100
	vnpAmount := amount * 100

	v := url.Values{}
	v.Set("vnp_Version", "2.1.0")
	v.Set("vnp_Command", "pay")
	v.Set("vnp_TmnCode", c.tmnCode)
	v.Set("vnp_Amount", fmt.Sprintf("%d", vnpAmount))
	v.Set("vnp_CreateDate", createDate)
	v.Set("vnp_CurrCode", "VND")
	v.Set("vnp_IpAddr", ipAddr)
	v.Set("vnp_Locale", "vn")
	v.Set("vnp_OrderInfo", desc)
	v.Set("vnp_OrderType", "other")
	v.Set("vnp_ReturnUrl", c.returnURL)
	v.Set("vnp_TxnRef", txnRef)

	// Sort keys alphabetically
	keys := make([]string, 0, len(v))
	for k := range v {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	var queryBuilder strings.Builder
	for i, k := range keys {
		if i > 0 {
			queryBuilder.WriteString("&")
		}
		// QueryEscape encodes space as + so we need to encode according to VNPay
		queryBuilder.WriteString(fmt.Sprintf("%s=%s", k, url.QueryEscape(v.Get(k))))
	}

	queryString := queryBuilder.String()
	// Do NOT replace "+" with "%20" because VNPay's backend expects space to be "+" (standard URLEncoder.encode behavior)
	// queryString = strings.ReplaceAll(queryString, "+", "%20")

	// Compute secure hash
	mac := hmac.New(sha512.New, []byte(c.hashSecret))
	mac.Write([]byte(queryString))
	secureHash := hex.EncodeToString(mac.Sum(nil))

	return fmt.Sprintf("%s?%s&vnp_SecureHash=%s", c.paymentURL, queryString, secureHash)
}

// VerifyChecksum checks the validity of VNPay callback parameters
func (c *VNPayClient) VerifyChecksum(params map[string][]string) bool {
	vnpSecureHash := ""
	cleanParams := make(map[string]string)

	for k, vals := range params {
		if len(vals) == 0 {
			continue
		}
		val := vals[0]
		if k == "vnp_SecureHash" {
			vnpSecureHash = val
			continue
		}
		if k == "vnp_SecureHashType" {
			continue
		}
		cleanParams[k] = val
	}

	// Sort keys alphabetically
	keys := make([]string, 0, len(cleanParams))
	for k := range cleanParams {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	var queryBuilder strings.Builder
	for i, k := range keys {
		if i > 0 {
			queryBuilder.WriteString("&")
		}
		queryBuilder.WriteString(fmt.Sprintf("%s=%s", k, url.QueryEscape(cleanParams[k])))
	}

	queryString := queryBuilder.String()
	// Do NOT replace "+" with "%20" to match GeneratePaymentURL change
	// queryString = strings.ReplaceAll(queryString, "+", "%20")

	mac := hmac.New(sha512.New, []byte(c.hashSecret))
	mac.Write([]byte(queryString))
	computedHash := hex.EncodeToString(mac.Sum(nil))

	return strings.ToLower(computedHash) == strings.ToLower(vnpSecureHash)
}
