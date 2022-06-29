package aggregation

import (
	"mysslee_qcloud/app/backend/db"

	"github.com/sirupsen/logrus"
)

type AggregateType struct {
	DomainId   int  //域名id
	AccountId  int  //用户id
	FromDomain bool //使用域名的标志
}

//用于接收重组信息

//聚合数据的处理者
type AggregateDashboardHandler struct {
	//用于接收
	acceptChan chan *AggregateType
	//用于做聚合处理
	calChan chan int
}

var AggrHandler *AggregateDashboardHandler

func init() {
	AggrHandler = &AggregateDashboardHandler{
		acceptChan: make(chan *AggregateType, 100),
		calChan:    make(chan int, 100),
	}
	go AggrHandler.FindUsersByDomainId()
	go AggrHandler.CalDashboardInfo()
}

func (c *AggregateDashboardHandler) SendAggregateRequest(aggregation *AggregateType) {
	c.acceptChan <- aggregation
}

//开go协程
func (c *AggregateDashboardHandler) FindUsersByDomainId() {
	for aggregationType := range c.acceptChan {

		if aggregationType.FromDomain { //如果是需要通过域名，聚合所有关注该域名的用户的数据
			err := c.findAndCalByDomainId(aggregationType.DomainId)
			if err != nil {
				logrus.Error("findAndCalByDomainId ", err)
			}
		} else { //用户添加或删除域名聚合用户的数据
			c.calChan <- aggregationType.AccountId
		}

	}
}

//查询所有关注的用户
func (c *AggregateDashboardHandler) findAndCalByDomainId(domainId int) (err error) {
	relations, err := db.GetUsersByDomainId(domainId)
	if err != nil {
		return err
	}
	for _, relation := range relations {
		c.calChan <- relation.AccountId
	}
	return nil
}

//开go协程
func (c *AggregateDashboardHandler) CalDashboardInfo() {
	for accountId := range c.calChan { //进行数据聚合
		err := CalDashboardInfo(accountId)
		if err != nil {
			logrus.Error("CalDashboardInfo ", err)
		}
	}
}
