package redeem

import (
	"github.com/gin-gonic/gin"
	"offercat/v0/internal/auth/model"
	"offercat/v0/internal/db"
	"offercat/v0/internal/lib"
	"strconv"
	"strings"
	"time"
)

type VerifyCodeRequest struct {
	Code string `json:"code" binding:"required"`
}

func VerifyCode(c *gin.Context) {
	uid := lib.Uid(c)
	var req VerifyCodeRequest
	var user model.User
	if err := c.ShouldBindJSON(&req); err != nil {
		lib.Err(c, 400, "参数错误", err)
		return
	}
	if err := db.DB.Where("id = ?", uid).First(&user).Error; err != nil {
		lib.Err(c, 500, "用户不存在", err)
		return
	}
	var code RedeemCode
	if err := db.DB.Where("code = ?", req.Code).First(&code).Error; err != nil {
		lib.Err(c, 500, "激活码不存在，核验激活码失败", err)
		return
	}
	if code.MaxUseCount <= code.UsedCount {
		lib.Err(c, 500, "该激活码已达到使用人数上限", nil)
		return
	}
	if code.ValidTo.Before(time.Now()) {
		lib.Err(c, 500, "激活码已过期", nil)
		return
	}
	if strings.HasPrefix(code.Tag, "vip") {
		split := strings.Split(code.Tag, ":")
		if len(split) != 2 {
			lib.Err(c, 500, "激活码在创建时出错，请联系管理员", nil)
		}
		// 把vip:1d中的1提取出来
		number, err := strconv.Atoi(strings.Trim(split[1], "d"))
		if err != nil {
			lib.Err(c, 500, "激活码在创建时出错，请联系管理员", nil)
		}

		if strings.Contains(code.UserID, strconv.Itoa(uid)) {
			lib.Err(c, 500, "该用户已使用过该激活码", nil)
			return
		}
		code.UsedCount++
		code.UserID += strconv.Itoa(uid) + ","
		db.DB.Save(&code)

		if user.VipExpireAt.IsZero() || (user.VipExpireAt).Before(time.Now()) {
			user.VipExpireAt = time.Now().AddDate(0, 0, number)
		} else {
			user.VipExpireAt = user.VipExpireAt.AddDate(0, 0, number)
		}

		db.DB.Save(&user)
		lib.Ok(c, "激活成功，VIP到期时间延长了"+strconv.Itoa(number)+"天", gin.H{
			"expire_at": user.VipExpireAt,
			"username":  user.Username,
			"tag":       code.Tag,
		})
	} else if strings.HasPrefix(code.Tag, "interviewPoint") {
		split := strings.Split(code.Tag, ":")
		if len(split) != 2 {
			lib.Err(c, 500, "激活码在创建时出错，请联系管理员", nil)
		}
		// 把interviewPoint:1h中的1提取出来
		hourNumber, err := strconv.Atoi(strings.Trim(split[1], "h"))
		// 把小时转换成分钟
		number := hourNumber * 60
		if err != nil {
			lib.Err(c, 500, "激活码在创建时出错，请联系管理员", nil)
		}

		if strings.Contains(code.UserID, strconv.Itoa(uid)) {
			lib.Err(c, 500, "该用户已使用过该激活码", nil)
			return
		}
		code.UsedCount++
		code.UserID += strconv.Itoa(uid) + ","
		db.DB.Save(&code)

		user.InterviewPoint += number

		db.DB.Save(&user)
		lib.Ok(c, "激活成功，面试点数增加了"+strconv.Itoa(number)+"分钟", gin.H{
			"point":    user.InterviewPoint,
			"username": user.Username,
			"tag":      code.Tag,
		})
	} else {
		lib.Err(c, 500, "激活码类型错误", nil)
	}

}
