package apis

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-admin-team/go-admin-core/sdk/api"
	_ "github.com/go-admin-team/go-admin-core/sdk/pkg/response"

	"go-admin/app/sn/models"
	"go-admin/app/sn/service"
	"go-admin/app/sn/service/dto"
)

type BoxRelationInfo struct {
	api.Api
}

func (e BoxRelationInfo) GetBoxRelationInfoList(c *gin.Context) {

	s := service.BatchInfo{}
	req := dto.BoxRelationInfoPageReq{}
	err := e.MakeContext(c).
		MakeOrm().
		Bind(&req, binding.Form).
		MakeService(&s.Service).
		Errors
	if err != nil {
		e.Logger.Error(err)
		e.Error(500, err, err.Error())
		return
	}

	list := make([]models.SNBoxRelation, 0)
	var count int64

	e.Logger.Info("BoxRelationInfoPageReq : ", req)

	if req.BoxId == 0 { //装箱操作台，如果没有箱号，不返回列表信息

	} else {
		err = s.GetRelationBoxInfoList(&req, &list, &count)
		if err != nil {
			e.Error(500, err, "查询失败")
			return
		}
	}

	e.PageOK(list, int(count), req.GetPageIndex(), req.GetPageSize(), "查询成功")
}

func (e BoxRelationInfo) AddBox(c *gin.Context) {
	s := service.BatchInfo{}
	req := dto.SNInfoBoxReq{}
	err := e.MakeContext(c).
		MakeOrm().
		Bind(&req, binding.JSON, nil).
		MakeService(&s.Service).
		Errors
	if err != nil {
		e.Logger.Error(err)
		e.Error(500, err, err.Error())
		return
	}

	ip1 := c.Request.Header.Get("X-Forward-For")
	ip2 := c.Request.Header.Get("X-Real-IP")
	ip3 := c.Request.Header.Get("REMOTE-HOST")
	e.Logger.Info("ip1:", ip1, " ip2=", ip2, " ip3=", ip3)

	req.ScanSource = ip2 //扫码枪来源通过IP来识别

	resultObj := dto.BoxRelationInfoResultObj{}

	//如果SN码，不在列表中，报错
	var snList []models.SNInfo
	e.Orm.Where("sn_code= ?", req.SNCode).Find(&snList)

	e.Logger.Info("AddBox snList: ", &snList)

	if len(snList) < 1 {
		resultObj.Status = 1 //
		e.OK(resultObj, "SN无效，不存在SN码")
		return
	}

	//SN码已经重复装箱，禁止重复操作
	var snRelationList []models.SNBoxRelation
	e.Orm.Where("sn_code= ?", req.SNCode).Find(&snRelationList)

	e.Logger.Info("AddBox snRelationList : ", &snRelationList)

	if len(snRelationList) > 0 {
		resultObj.Status = 2 //
		e.OK(resultObj, "SN码已经装箱，禁止重复操作")
		return
	}

	err = s.SNPackBox(&req)
	if err != nil {
		e.Error(500, err, fmt.Sprintf("SN装箱失败！错误详情：%s", err.Error()))
		return
	}

	var snBoxRelation = models.SNBoxRelation{}
	var whereStr string
	whereStr = "sn_code='" + req.SNCode + "'"
	e.Orm.First(&snBoxRelation, whereStr)
	resultObj.Status = 0 //装箱成功
	resultObj.BoxId = snBoxRelation.BoxId

	e.Logger.Info("-----------------------")

	bSum := 10
	var listBoxRelation []models.SNBoxRelation
	e.Orm.Where("box_id= ?", snBoxRelation.BoxId).Find(&listBoxRelation)

	e.Logger.Info("-----------------AddBox listBoxRelation boxId: ", snBoxRelation.BoxId, ", bSum:", bSum, ", len:", len(listBoxRelation))
	if len(listBoxRelation) == bSum { //表示箱子刚好装满
		resultObj.Status = 4 //装满一箱
		resultObj.BoxSNCodeList = listBoxRelation
	}

	e.OK(resultObj, "更新成功")
}
