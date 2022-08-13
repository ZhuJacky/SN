package apis

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-admin-team/go-admin-core/sdk/api"
	"github.com/go-admin-team/go-admin-core/sdk/pkg/jwtauth/user"
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

	err = s.GetRelationBoxInfoList(&req, &list, &count)
	if err != nil {
		e.Error(500, err, "查询失败")
		return
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

	req.SetUpdateBy(user.GetUserId(c))
	req.ScanSource = "777"
	req.Status = 3
	e.Logger.Info("AddBox PackBox:", req)

	if req.SNCode == "010695174052138310202208007218930500E056013" {
		e.OK(req, "SN无效")
		return

	}

	err = s.SNPackBox(&req)
	if err != nil {
		e.Error(500, err, fmt.Sprintf("SN装箱失败！错误详情：%s", err.Error()))
		return
	}
	e.OK(req.GetSNCode(), "更新成功")
}
