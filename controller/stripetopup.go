package controller

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/stripe/stripe-go"
	"github.com/stripe/stripe-go/webhook"
	"one-api/common"
	"one-api/model"
)

var (
	endpointSecret = os.Getenv("STRIPE_WEBHOOK_SECRET")
)

func StripeWebhook(c *gin.Context) {
	const MaxBodyBytes = int64(65536)
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, MaxBodyBytes)

	payload, err := c.GetRawData()
	if err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"message": "error", "data": "无法读取请求体"})
		return
	}

	event, err := webhook.ConstructEvent(payload, c.GetHeader("Stripe-Signature"), endpointSecret)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "error", "data": "签名验证失败"})
		return
	}

	switch event.Type {
	case "payment_intent.succeeded":
		var paymentIntent stripe.PaymentIntent
		err := json.Unmarshal(event.Data.Raw, &paymentIntent)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"message": "error", "data": "解析支付意图失败"})
			return
		}

		tradeNo := paymentIntent.Metadata["trade_no"]
		LockOrder(tradeNo)
		defer UnlockOrder(tradeNo)

		topUp := model.GetTopUpByTradeNo(tradeNo)
		if topUp == nil {
			log.Printf("Stripe 回调未找到订单: %v", tradeNo)
			return
		}

		if topUp.Status == "pending" {
			topUp.Status = "success"
			err := topUp.Update()
			if err != nil {
				log.Printf("Stripe 回调更新订单失败: %v", topUp)
				return
			}
			err = model.IncreaseUserQuota(topUp.UserId, topUp.Amount*int(common.QuotaPerUnit))
			if err != nil {
				log.Printf("Stripe 回调更新用户失败: %v", topUp)
				return
			}
			log.Printf("Stripe 回调更新用户成功 %v", topUp)
			model.RecordLog(topUp.UserId, model.LogTypeTopup, fmt.Sprintf("使用Stripe充值成功，充值金额: %v，支付金额：%f", common.LogQuota(topUp.Amount*int(common.QuotaPerUnit)), topUp.Money))
		}
	case "payment_intent.payment_failed":
		log.Printf("支付失败: %v", event)
	}

	c.JSON(http.StatusOK, gin.H{"message": "success"})
}
