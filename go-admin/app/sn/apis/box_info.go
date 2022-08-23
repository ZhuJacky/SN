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

type BoxInfo struct {
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
func (e BoxInfo) GetBoxInfoList(c *gin.Context) {

	s := service.BatchInfo{}
	req := dto.BoxInfoPageReq{}
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

	list := make([]models.SNBoxInfo, 0)
	var count int64

	err = s.GetBoxInfoList(&req, &list, &count)
	if err != nil {
		e.Error(500, err, "查询失败")
		return
	}

	e.PageOK(list, int(count), req.GetPageIndex(), req.GetPageSize(), "查询成功")
}

func (e BoxInfo) UpdateBoxSum(c *gin.Context) {
	s := service.BatchInfo{}
	req := dto.BoxInfoUpdateReq{}
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
	e.Logger.Info("UpdateBoxSum:", req)

	err = s.UpdateBoxSum(&req)
	if err != nil {
		e.Error(500, err, fmt.Sprintf("装箱数量更新失败！错误详情：%s", err.Error()))
		return
	}
	e.OK(req.GetId(), "更新成功")
}

func (e BoxInfo) GetExWarehouseBoxList(c *gin.Context) {

	s := service.BatchInfo{}
	req := dto.BoxInfoPageReq{}
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

	list := make([]models.SNBoxInfo, 0)
	var count int64
	req.Status = 2 //只是查询出库列表

	err = s.GetBoxInfoList(&req, &list, &count)
	if err != nil {
		e.Error(500, err, "查询失败")
		return
	}

	e.PageOK(list, int(count), req.GetPageIndex(), req.GetPageSize(), "查询成功")
}

func (e BoxInfo) UpdateExWarehouseBoxStatus(c *gin.Context) {
	s := service.BatchInfo{}
	req := dto.BoxInfoUpdateStatusReq{}
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
	e.Logger.Info("UpdateExWarehouseBoxStatus:", req)

	//如果箱号码，不在列表中，报错
	var boxList []models.SNBoxInfo
	e.Orm.Where("box_id= ?", req.BoxId).Find(&boxList)

	e.Logger.Info("UpdateExWarehouseBoxStatus boxList: ", &boxList)

	resultObj := dto.BoxInfoResultObj{}

	if len(boxList) < 1 {
		resultObj.BoxId = req.BoxId
		resultObj.Status = -1
		e.OK(resultObj, "箱号无效，不存在箱号信息")
		return
	}

	err = s.UpdateBoxStatus(&req)
	if err != nil {
		e.Error(500, err, fmt.Sprintf("出库失败！错误详情：%s", err.Error()))
		return
	}

	resultObj.BoxId = req.BoxId
	resultObj.Status = 0
	e.OK(resultObj, "更新成功")
}

func (e BoxInfo) GetEnWarehouseBoxList(c *gin.Context) {

	s := service.BatchInfo{}
	req := dto.BoxInfoPageReq{}
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

	list := make([]models.SNBoxInfo, 0)
	var count int64
	req.Status = 1 //只是查询入库列表

	err = s.GetBoxInfoList(&req, &list, &count)
	if err != nil {
		e.Error(500, err, "查询失败")
		return
	}

	e.PageOK(list, int(count), req.GetPageIndex(), req.GetPageSize(), "查询成功")
}
