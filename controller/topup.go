package controller

import "C"
import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/stripe/stripe-go/v76"
	"github.com/stripe/stripe-go/v76/checkout/session"
	"log"
	"one-api/common"
	"one-api/model"
	"strconv"
	"strings"
	"time"
)

type PayRequest struct {
	Amount        int    `json:"amount"`
	PaymentMethod string `json:"payment_method"`
	TopUpCode     string `json:"top_up_code"`
}

type AmountRequest struct {
	Amount    int    `json:"amount"`
	TopUpCode string `json:"top_up_code"`
}

func genStripeLink(referenceId string, customerId string, email string, amount int64) (string, error) {
	if !strings.HasPrefix(common.StripeApiSecret, "sk_") {
		return "", fmt.Errorf("无效的Stripe API密钥")
	}

	stripe.Key = common.StripeApiSecret

	params := &stripe.CheckoutSessionParams{
		ClientReferenceID: stripe.String(referenceId),
		SuccessURL:        stripe.String(common.ServerAddress + "/log"),
		CancelURL:         stripe.String(common.ServerAddress + "/topup"),
		LineItems: []*stripe.CheckoutSessionLineItemParams{
			{
				Price:    stripe.String(common.StripePriceId),
				Quantity: stripe.Int64(amount),
			},
		},
		Mode: stripe.String(string(stripe.CheckoutSessionModePayment)),
	}

	if "" == customerId {
		if "" != email {
			params.CustomerEmail = stripe.String(email)
		}

		params.CustomerCreation = stripe.String(string(stripe.CheckoutSessionCustomerCreationAlways))
	} else {
		params.Customer = stripe.String(customerId)
	}

	result, err := session.New(params)
	if err != nil {
		return "", err
	}

	return result.URL, nil
}

func GetPayAmount(count float64) float64 {
	return count * common.StripeUnitPrice
}

func GetChargedAmount(count float64, user model.User) float64 {
	topUpGroupRatio := common.GetTopupGroupRatio(user.Group)
	if topUpGroupRatio == 0 {
		topUpGroupRatio = 1
	}

	return count * topUpGroupRatio
}

func RequestPayLink(c *gin.Context) {
	var req PayRequest
	err := c.ShouldBindJSON(&req)
	if err != nil {
		c.JSON(200, gin.H{"message": err.Error(), "data": 10})
		return
	}
	if !common.PaymentEnabled {
		c.JSON(200, gin.H{"message": "error", "data": "管理员未开启在线支付"})
		return
	}
	if req.PaymentMethod != "stripe" {
		c.JSON(200, gin.H{"message": "error", "data": "不支持的支付渠道"})
		return
	}
	if req.Amount < common.MinTopUp {
		c.JSON(200, gin.H{"message": fmt.Sprintf("充值数量不能小于 %d", common.MinTopUp), "data": 10})
		return
	}
	if req.Amount > 10000 {
		c.JSON(200, gin.H{"message": "充值数量不能大于 10000", "data": 10})
		return
	}

	id := c.GetInt("id")
	user, _ := model.GetUserById(id, false)
	chargedMoney := GetChargedAmount(float64(req.Amount), *user)

	reference := fmt.Sprintf("new-api-ref-%d-%d-%s", user.Id, time.Now().UnixMilli(), common.RandomString(4))
	referenceId := "ref_" + common.Sha1(reference)

	payLink, err := genStripeLink(referenceId, user.StripeCustomer, user.Email, int64(req.Amount))
	if err != nil {
		log.Println("获取Stripe Checkout支付链接失败", err)
		c.JSON(200, gin.H{"message": "error", "data": "拉起支付失败"})
		return
	}

	topUp := &model.TopUp{
		UserId:     id,
		Amount:     req.Amount,
		Money:      chargedMoney,
		TradeNo:    referenceId,
		CreateTime: time.Now().Unix(),
		Status:     common.TopUpStatusPending,
	}
	err = topUp.Insert()
	if err != nil {
		c.JSON(200, gin.H{"message": "error", "data": "创建订单失败"})
		return
	}
	c.JSON(200, gin.H{
		"message": "success",
		"data": gin.H{
			"payLink": payLink,
		},
	})
}

func RequestAmount(c *gin.Context) {
	var req AmountRequest
	err := c.ShouldBindJSON(&req)
	if err != nil {
		c.JSON(200, gin.H{"message": "error", "data": "参数错误"})
		return
	}
	if !common.PaymentEnabled {
		c.JSON(200, gin.H{"message": "error", "data": "管理员未开启在线支付"})
		return
	}
	if req.Amount < common.MinTopUp {
		c.JSON(200, gin.H{"message": "error", "data": fmt.Sprintf("充值数量不能小于 %d", common.MinTopUp)})
		return
	}
	id := c.GetInt("id")
	user, _ := model.GetUserById(id, false)
	payMoney := GetPayAmount(float64(req.Amount))
	chargedMoney := GetChargedAmount(float64(req.Amount), *user)
	c.JSON(200, gin.H{
		"message": "success",
		"data": gin.H{
			"payAmount":     strconv.FormatFloat(payMoney, 'f', 2, 64),
			"chargedAmount": strconv.FormatFloat(chargedMoney, 'f', 2, 64),
		},
	})
}
