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

type SNInfo struct {
	api.Api
}

// GetPage
// @Summary 岗位列表数据
// @Description 获取JSON
// @Tags 岗位
// @Param postName query string false "postName"
// @Param postCode query string false "postCode"
// @Param postId query string false "postId"
// @Param status query string false "status"
// @Success 200 {object} response.Response "{"code": 200, "data": [...]}"
// @Router /api/v1/post [get]
// @Security Bearer
func (e SNInfo) GetSNInfoList(c *gin.Context) {

	s := service.BatchInfo{}
	req := dto.SNInfoPageReq{}
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

	list := make([]models.SNInfo, 0)
	var count int64

	err = s.GetSNInfoList(&req, &list, &count)
	if err != nil {
		e.Error(500, err, "查询失败")
		return
	}

	e.PageOK(list, int(count), req.GetPageIndex(), req.GetPageSize(), "查询成功")
}

func (e SNInfo) UpdateStatus(c *gin.Context) {
	s := service.BatchInfo{}
	req := dto.SNInfoUpdateReq{}
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
	e.Logger.Info("UpdateStatus:", req)

	err = s.UpdateSNInfoStatus(&req)
	if err != nil {
		e.Error(500, err, fmt.Sprintf("SN状态更新失败！错误详情：%s", err.Error()))
		return
	}
	e.OK(req.GetId(), "更新成功")
}

func (e SNInfo) PackBox(c *gin.Context) {
	s := service.BatchInfo{}
	req := dto.SNInfoPackBoxReq{}
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
	e.Logger.Info("PackBox:", req)

	err = s.SNPackBox(&req)
	if err != nil {
		e.Error(500, err, fmt.Sprintf("SN装箱失败！错误详情：%s", err.Error()))
		return
	}
	e.OK(req.GetSNCode(), "更新成功")
}
