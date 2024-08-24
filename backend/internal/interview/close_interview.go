package interview

import (
	"github.com/gin-gonic/gin"
	"offercat/v0/internal/auth/model"
	"offercat/v0/internal/db"
	"offercat/v0/internal/interview/common"
	"offercat/v0/internal/lib"
	"strconv"
	"time"
)

type CloseInterviewRequest struct {
	InterviewID uint `json:"interview_id" binding:"required"`
}

func CloseInterview(c *gin.Context) {
	var req CloseInterviewRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		lib.Err(c, 400, "解析请求体失败", err)
		return
	}
	var user model.User
	uid := lib.Uid(c)
	if err := db.DB.Where("id = ?", uid).First(&user).Error; err != nil {
		lib.Err(c, 400, "用户不存在", err)
		return
	}
	var interview common.Interview
	if err := db.DB.Where("id = ? AND user_id = ?", req.InterviewID, uid).First(&interview).Error; err != nil {
		lib.Err(c, 400, "面试不存在", err)
		return
	}
	if interview.Closed == true {
		lib.Err(c, 400, "面试处于关闭状态，请勿重复关闭面试", nil)
		return
	}
	//关闭面试
	if err := db.DB.Model(&interview).Update("closed", true).Error; err != nil {
		lib.Err(c, 400, "关闭面试失败", err)
		return
	}
	//获取上下文中的"cost_type"，判断用户使用的是vip还是面试点数
	costType, _ := c.Get("cost_type")
	if costType == "vip" {
		//如果是vip，不扣除面试点数
		lib.Ok(c, "关闭面试成功", gin.H{
			"cost_type":  costType,
			"time_spent": strconv.Itoa(int(time.Now().Sub(interview.StartTime).Minutes())) + "分钟",
		})
	} else {
		//如果是面试点数，扣除面试点数
		//先计算面试消耗的时间，也就是点数
		timeCost := time.Now().Sub(interview.StartTime)
		//计算点数
		user.InterviewPoint = user.InterviewPoint - int(timeCost.Minutes())
		if err := db.DB.Model(&user).Update("interview_point", user.InterviewPoint).Error; err != nil {
			lib.Err(c, 400, "扣除面试点数失败", err)
			//如果扣除面试点数失败，将面试状态改回未关闭
			db.DB.Model(&interview).Update("closed", false)
			return
		}
		lib.Ok(c, "关闭面试成功", gin.H{
			"cost_type":  costType,
			"cost_point": int(timeCost.Minutes()),
			"left_point": user.InterviewPoint,
			"time_spent": strconv.Itoa(int(timeCost.Minutes())) + "分钟",
		})
	}

}
