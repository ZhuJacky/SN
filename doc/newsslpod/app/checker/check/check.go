// Package check provides ...
package check

import (
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"mysslee_qcloud/app/checker/prom"
	"mysslee_qcloud/config"
	"mysslee_qcloud/model"
)

var (
	DomainChecker *Checker

	fastTasks     = int32(config.Conf.Checker.Task.FastWorker * config.NumCPU)
	fullTasks     = int32(config.Conf.Checker.Task.FullWorker * config.NumCPU)
	kafkaTasksNum = int32(config.Conf.Checker.Task.FastWorker * config.NumCPU)
)

func init() {
	// 设计思考：
	// 1. 预估有 25w 域名（包括多IP）
	// 2. 有 5 台checker节点
	// 3. 快速检测平均每域名5s
	// 4. 全量检测平均每域名40s
	//
	// 快速检测：
	// 1. 每台机器每10m需检测：40w/5=5w
	// 2. 每台机器需要同时处理任务：5w/10/60*5=420
	// 全量检测：
	// 1. 每台机器每24h需要处理：25w/5=5w
	// 2. 每台机器需要同时处理任务：5w/24/60/60*40=24
	// 按实际情况，处理能力上下浮动25%。
	// 同时间接纳任务数：1450
	//
	// 每个检测5k*1w=490m，即需要内存为512m的机器配置
	// CPU 至少两核
	DomainChecker = &Checker{}
	for i := int32(0); i <= fastTasks; i++ { // 600*5=3000
		w := &fastCheckWorker{
			workChan: make(chan *model.DomainResult, 10),
		}
		go w.do()
		DomainChecker.fastWorkers = append(DomainChecker.fastWorkers, w)
	}
	for i := int32(0); i <= fullTasks; i++ { // (40/5)*600=4800
		w2 := &fullCheckWorker{
			workChan: make(chan *model.DomainResult, 103),
		}
		go w2.do()
		DomainChecker.fullWorkers = append(DomainChecker.fullWorkers, w2)
	}
	for i := int32(0); i <= kafkaTasksNum; i++ { // 处理kafka任务 沿用快速检测
		w3 := &kafkaCheckWorker{
			workChan: make(chan *model.KafkaDomainInfo, 10),
		}
		go w3.do()
		DomainChecker.kafkaWorkers = append(DomainChecker.kafkaWorkers, w3)
	}
}

var (
	fastIndex  int32 = 0
	fullIndex  int32 = 0
	kafkaIndex int32 = 0
)

// 检测器
type Checker struct {
	lock         sync.Mutex
	fastWorkers  []*fastCheckWorker
	fullWorkers  []*fullCheckWorker
	kafkaWorkers []*kafkaCheckWorker
}

func (c *Checker) DoFast(dr model.DomainResult) {
	worker := c.idleFastWorker(0)

	prom.PromFastDetection.WithLabelValues("total").Inc()

	worker.workChan <- &dr
	worker.incr()
}

func (c *Checker) DoFull(dr *model.DomainResult) {
	worker := c.idleFullWorker(0)

	prom.PromFullDetection.WithLabelValues("total").Inc()

	worker.incr()
	worker.workChan <- dr
}

func (c *Checker) DoKafka(dr model.KafkaDomainInfo) {
	worker := c.idleKafkaWorker(0)

	worker.incr()
	worker.workChan <- &dr
}

func (c *Checker) idleFastWorker(count int) *fastCheckWorker {
	if !atomic.CompareAndSwapInt32(&fastIndex, fastTasks, 0) {
		atomic.AddInt32(&fastIndex, 1)
	}
	worker := c.fastWorkers[fastIndex]
	if worker.busy < 5 || count > 5 {
		return worker
	}

	count++
	time.Sleep(time.Millisecond * 10)
	return c.idleFastWorker(count)
}

func (c *Checker) idleFullWorker(count int) *fullCheckWorker {
	if !atomic.CompareAndSwapInt32(&fullIndex, fullTasks, 0) {
		atomic.AddInt32(&fullIndex, 1)
	}
	worker := c.fullWorkers[fullIndex]
	if worker.busy < 103 || count > 5 {
		return worker
	}

	count++
	time.Sleep(time.Millisecond * 10)
	return c.idleFullWorker(count)
}

func (c *Checker) idleKafkaWorker(count int) *kafkaCheckWorker {
	if !atomic.CompareAndSwapInt32(&kafkaIndex, kafkaTasksNum, 0) {
		atomic.AddInt32(&kafkaIndex, 1)
	}
	worker := c.kafkaWorkers[kafkaIndex]
	if worker.busy < 5 || count > 5 {
		return worker
	}

	count++
	time.Sleep(time.Millisecond * 10)
	return c.idleKafkaWorker(count)
}

// PrintWorkerBusy 打印各个workder的busy数
func (c *Checker) PrintWorkerBusy() {
	var s = []int32{}
	for _, i := range c.kafkaWorkers {
		s = append(s, i.busy)
	}
	fmt.Println(s)
}
