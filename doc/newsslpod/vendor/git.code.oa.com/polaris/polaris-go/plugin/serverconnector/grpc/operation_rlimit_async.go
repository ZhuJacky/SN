package grpc

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"git.code.oa.com/polaris/polaris-go/pkg/clock"
	"git.code.oa.com/polaris/polaris-go/pkg/config"
	"git.code.oa.com/polaris/polaris-go/pkg/flow/quota"
	"git.code.oa.com/polaris/polaris-go/pkg/log"
	"git.code.oa.com/polaris/polaris-go/pkg/model"
	rlimitpb "git.code.oa.com/polaris/polaris-go/pkg/model/pb/metric"
	"git.code.oa.com/polaris/polaris-go/pkg/network"
	connector "git.code.oa.com/polaris/polaris-go/plugin/serverconnector/common"
	"github.com/golang/protobuf/jsonpb"
	"sync"
	"sync/atomic"
	"time"
	//grpc "google.golang.org/grpc"
	"github.com/golang/protobuf/proto"
)

type SendRequest struct {
	reqType string
	request proto.Message
	stream  *StreamWrapper
	win     *quota.RateLimitWindow
}

type StreamWrapper struct {
	// ip:port
	streamKey string
	op        string

	conn *network.Connection

	rLimitClient  rlimitpb.RateLimitGRPCClient
	acquireClient rlimitpb.RateLimitGRPC_AcquireQuotaClient

	metricClient       rlimitpb.MetricGRPCClient
	metricReportClient rlimitpb.MetricGRPC_ReportClient

	receiveCtrlChan chan int
	receiveRunning  int32
	cancel          context.CancelFunc

	lastUseTime int64
	canceled    int32
	once        sync.Once

	ref    int64
	keyRef int64

	lock sync.Mutex
}

func (c *StreamWrapper) acquire() {
	atomic.AddInt64(&c.ref, 1)
}

func (c *StreamWrapper) release() {
	atomic.AddInt64(&c.ref, -1)
}

func (c *StreamWrapper) keyAcquire() {
	atomic.AddInt64(&c.keyRef, 1)
}

func (c *StreamWrapper) keyRelease() {
	atomic.AddInt64(&c.keyRef, -1)
}

// 返回是否需要重新创建连接
func (c *StreamWrapper) handleNetWorkError(err error, runType string) bool {
	return true
}

// 设置stream连接
func (c *StreamWrapper) SetStream(rConnector *AsyncRateLimitConnector) error {
	hashKey := make([]byte, 0, len(c.streamKey))
	buf := bytes.NewBuffer(hashKey)
	buf.WriteString(c.streamKey)

	addr, instance, err := rConnector.getKeyHashAddr(string(buf.Bytes()))
	if err != nil {
		return err
	}
	conn, err := rConnector.connManager.ConnectByAddr(config.MetricCluster, addr, instance)
	if err != nil {
		return err
	}
	c.conn = conn
	if c.op == connector.OpKeyRateLimitAcquire {
		log.GetBaseLogger().Debugf("[CheckConnection]traceCheck key:%s addr:%s", c.streamKey,
			c.conn.ConnID.Address)
	}
	ctx := context.Background()
	streamCtx, cancel := context.WithCancel(ctx)
	c.cancel = cancel
	switch c.op {
	case connector.OpKeyRateLimitAcquire:
		c.rLimitClient = rlimitpb.NewRateLimitGRPCClient(network.ToGRPCConn(conn.Conn))
		c.acquireClient, err = c.rLimitClient.AcquireQuota(streamCtx)
	}
	atomic.StoreInt32(&c.canceled, 0)
	if err != nil {
		return err
	}
	return nil
}

// stop
func (c *StreamWrapper) stop() error {
	if atomic.LoadInt32(&c.canceled) == 1 {
		return errors.New("has stopped")
	}
	c.lock.Lock()
	defer c.lock.Unlock()
	if atomic.LoadInt32(&c.canceled) == 1 {
		return errors.New("has stopped")
	}
	log.GetBaseLogger().Debugf("[CheckConnection]traceCheck stop key:%s addr:%s", c.streamKey,
		c.conn.ConnID.Address)
	atomic.StoreInt32(&c.canceled, 1)
	c.once.Do(func() {
		if c.op == connector.OpKeyRateLimitAcquire {
			c.acquireClient.CloseSend()
		}
		addr := c.conn.ConnID.Address
		closed := c.conn.ForceClose()
		if !closed {
			log.GetBaseLogger().Warnf("[CheckConnection]traceCheck close failed addr:%s", addr)
		}
	})
	c.cancel()
	c.receiveCtrlChan <- 1
	log.GetBaseLogger().Debugf("[CheckConnection]traceCheck stop key:%s addr:%s done", c.streamKey,
		c.conn.ConnID.Address)
	return nil
}

type WindowRecord struct {
	win      *quota.RateLimitWindow
	lastTime int64
}

type WindowInitJob struct {
	reqType string
	request proto.Message
	win     *quota.RateLimitWindow
}

// 目前只实现了 RateLimit-Acquire的异步 和 metric-report的异步
type AsyncRateLimitConnector struct {
	*rateLimitConnector

	sendRunning int32

	taskChan     chan *SendRequest
	initTaskChan chan *WindowInitJob

	acquireWindowMap  sync.Map
	acquireStreamMap  sync.Map
	acquireKeyAddrMap sync.Map

	rateLimitInitConnMap sync.Map

	start int32
	lock  *sync.Mutex

	destroyed int32

	ctx    context.Context
	cancel context.CancelFunc
}

// NewAsyncRateLimitConnector
func NewAsyncRateLimitConnector(connManager network.ConnectionManager) *AsyncRateLimitConnector {
	rLimitConn := &AsyncRateLimitConnector{}
	rLimitConn.rateLimitConnector = &rateLimitConnector{connManager: connManager}
	rLimitConn.connManager = connManager
	rLimitConn.taskChan = make(chan *SendRequest, 10000)
	rLimitConn.initTaskChan = make(chan *WindowInitJob, 2048)
	rLimitConn.start = 0
	rLimitConn.sendRunning = 0
	rLimitConn.lock = &sync.Mutex{}
	rLimitConn.ctx, rLimitConn.cancel = context.WithCancel(context.Background())
	return rLimitConn
}

// Start
func (c *AsyncRateLimitConnector) Start() error {
	go c.doSend()
	go c.DoWindowInit()
	go c.ClearExpireRecords()
	// for test
	//go PrintNum()
	atomic.StoreInt32(&c.start, 1)
	return nil
}

// Destroy
func (c *AsyncRateLimitConnector) Destroy() error {
	if atomic.LoadInt32(&c.destroyed) == 1 {
		return nil
	}
	c.cancel()
	atomic.StoreInt32(&c.destroyed, 1)
	c.acquireStreamMap.Range(func(key, value interface{}) bool {
		stream := value.(*StreamWrapper)
		if stream != nil {
			c.closeStream(stream, key.(string), true)
		}
		return true
	})
	c.rateLimitInitConnMap.Range(func(key, value interface{}) bool {
		s := value.(*StreamWrapper)
		s.conn.Release(connector.OpKeyRateLimitInit)
		c.rateLimitInitConnMap.Delete(key)
		return true
	})
	c.acquireWindowMap.Range(func(key, value interface{}) bool {
		c.acquireWindowMap.Delete(key)
		return true
	})
	return nil
}

// 清理过期的用于回调而保存的窗口记录
func (c *AsyncRateLimitConnector) ClearExpireRecords() {
	dur := time.Second * 3
	ticker := time.NewTicker(dur)
	defer ticker.Stop()
	for {
		if atomic.LoadInt32(&c.destroyed) == 1 {
			return
		}
		select {
		case <-ticker.C:
			timeNow := time.Now().Unix()
			c.acquireStreamMap.Range(func(key, value interface{}) bool {
				s := value.(*StreamWrapper)
				ks := key.(string)
				if timeNow-atomic.LoadInt64(&s.lastUseTime) > 60 {
					c.closeStream(s, ks, true)
				} else {
					log.GetBaseLogger().Debugf("[CheckConnection]traceCheck acquireStreamMap release %s keyRef "+
						"%d", ks, atomic.LoadInt64(&s.keyRef))
					if atomic.LoadInt64(&s.keyRef) <= 0 {
						c.closeStream(s, ks, false)
					}
				}
				return true
			})
			c.rateLimitInitConnMap.Range(func(key, value interface{}) bool {
				s := value.(*StreamWrapper)
				ks := key.(string)
				log.GetBaseLogger().Debugf("[CheckConnection]traceCheck rateLimitInitConnMap key: %s ", ks)
				if timeNow-atomic.LoadInt64(&s.lastUseTime) > 20 {
					s.conn.Release(connector.OpKeyRateLimitInit)
					c.rateLimitInitConnMap.Delete(key)
				}
				return true
			})
			c.acquireWindowMap.Range(func(key, value interface{}) bool {
				s := value.(*WindowRecord)
				if timeNow-s.lastTime > 3 {
					c.acquireWindowMap.Delete(key)
				}
				return true
			})
		case <-c.ctx.Done():
			log.GetBaseLogger().Infof("AsyncRateLimitConnector ClearExpireRecords exit")
			return
		}
	}
}

// 获取一个stream (一个serviceKey 对应一个stream)
func (c *AsyncRateLimitConnector) getStreamWrapper(streamKey string, op string) (*StreamWrapper, error) {
	var err error
	streamWrapper := &StreamWrapper{}
	streamWrapper.streamKey = streamKey
	streamWrapper.op = op
	streamWrapper.receiveCtrlChan = make(chan int, 3)
	streamWrapper.receiveRunning = 0
	streamWrapper.canceled = 0
	streamWrapper.lastUseTime = clock.GetClock().Now().Unix()
	err = streamWrapper.SetStream(c)
	if err != nil {
		return nil, err
	}
	return streamWrapper, nil
}

func (c *AsyncRateLimitConnector) closeStream(stream *StreamWrapper, mapKey string, force bool) {
	log.GetBaseLogger().Debugf("[CheckConnection]closeStream")
	if !force {
		cnt := 0
		for {
			if atomic.LoadInt64(&stream.ref) <= 0 {
				break
			} else {
				time.Sleep(time.Millisecond * 20)
				cnt++
				if cnt >= 10 {
					return
				}
			}
		}
	}
	err := stream.stop()
	if err != nil {
		log.GetBaseLogger().Warnf("[CheckConnection]closeStream error:%s", err.Error())
	} else {
		if stream.op == connector.OpKeyRateLimitAcquire {
			c.acquireStreamMap.Delete(mapKey)
		}
	}
}

func (c *AsyncRateLimitConnector) getKeyHashAddr(key string) (string, model.Instance, error) {
	hashKey := make([]byte, 0, len(key))
	buf := bytes.NewBuffer(hashKey)
	buf.WriteString(key)
	addr, instance, err := c.connManager.GetHashExpectedInstance(config.MetricCluster, buf.Bytes())
	if err != nil {
		return "", nil, err
	}
	if log.GetBaseLogger().IsLevelEnabled(log.DebugLog) && instance != nil {
		log.GetBaseLogger().Debugf("[CheckConnection]---getKeyHashAddr %s %s %s", buf.String(), addr,
			instance.GetHost())
	}
	return addr, instance, nil
}

func (c *AsyncRateLimitConnector) DoWindowInit() {
	for {
		if atomic.LoadInt32(&c.destroyed) == 1 {
			return
		}
		select {
		case initJob := <-c.initTaskChan:
			switch initJob.reqType {
			case connector.OpKeyRateLimitInit:
				err := c.doSyncRateLimitInit(initJob)
				if err != nil {
					log.GetBaseLogger().Warnf("[CheckConnection]traceCheck %s init error:%s", err.Error())
				}
			}
		case <-c.ctx.Done():
			log.GetBaseLogger().Infof("AsyncRateLimitConnector DoWindowInit exit")
			return
		}
	}
}

// doSend
func (c *AsyncRateLimitConnector) doSend() {
	atomic.StoreInt32(&c.sendRunning, 1)
	for {
		if atomic.LoadInt32(&c.destroyed) == 1 {
			log.GetBaseLogger().Infof("AsyncRateLimitConnector doSend exit")
			return
		}
		select {
		case <-c.ctx.Done():
			log.GetBaseLogger().Infof("AsyncRateLimitConnector doSend exit")
			atomic.StoreInt32(&c.sendRunning, 0)
			return
		case msg := <-c.taskChan:
			timeNow := clock.GetClock().Now().Unix()
			switch msg.reqType {
			case connector.OpKeyRateLimitAcquire:
				req := msg.request.(*rlimitpb.RateLimitRequest)
				if atomic.LoadInt32(&msg.stream.canceled) == 1 {
					msg.stream.release()
					continue
				}
				err := msg.stream.acquireClient.Send(req)
				if err != nil {
					log.GetBaseLogger().Warnf("AsyncRateLimitConnector  %s %s realSend error %s:",
						req.GetNamespace().GetValue(), req.GetService().GetValue(), err)
					c.closeStream(msg.stream, msg.stream.conn.Address, true)
				} else {
					if log.GetBaseLogger().IsLevelEnabled(log.DebugLog) {
						log.GetBaseLogger().Debugf("[RateLimit]do send ok ",
							time.Now().UnixNano()/int64(time.Millisecond), msg.stream.conn.Address, req.String())
					}
					atomic.StoreInt64(&msg.stream.lastUseTime, timeNow)
				}
				msg.stream.release()
			default:
				continue
			}
		}
	}
}

// 接收Acquire的回包
func (c *AsyncRateLimitConnector) doReceiveAcquire(streamWrapper *StreamWrapper) {
	atomic.StoreInt32(&streamWrapper.receiveRunning, 1)
	for {
		select {
		case stop := <-streamWrapper.receiveCtrlChan:
			_ = stop
			log.GetBaseLogger().Warnf("doReceiveAcquire stop")
			atomic.StoreInt32(&streamWrapper.receiveRunning, 0)
			return
		default:
			msg, err := streamWrapper.acquireClient.Recv()
			if err != nil {
				log.GetBaseLogger().Warnf("[RateLimit]RateLimit acquireClient receive msg err:%s", err.Error())
				continue
			}
			c.Report(streamWrapper)
			if log.GetBaseLogger().IsLevelEnabled(log.DebugLog) {
				log.GetBaseLogger().Debugf("[RateLimit]recv ok  ", msg.String())
			}
			mapValue, ok := c.acquireWindowMap.Load(msg.GetKey().GetValue())
			if !ok {
				log.GetBaseLogger().Warnf("[RateLimit]acquireRdMap.Load load rWindow not ok key:%s",
					msg.GetKey().GetValue())
				continue
			}
			record := mapValue.(*WindowRecord)
			if record.win != nil {
				record.win.OnRemoteAcquireResponse(msg)
			}
		}
	}
}

//上报
func (c *AsyncRateLimitConnector) Report(streamWrapper *StreamWrapper) {
	c.connManager.ReportSuccess(streamWrapper.conn.ConnID, 0, 10)
}

// sendToLocalChannel
func (c *AsyncRateLimitConnector) sendToLocalChannel(request *SendRequest) error {
	select {
	case c.taskChan <- request:
		return nil
	default:
		log.GetBaseLogger().Warnf("[RateLimit]AsyncRateLimitConnector localChannel full")
		return errors.New("buffer full")
	}
}

func (c *AsyncRateLimitConnector) sendToInitChannel(initJob *WindowInitJob) error {
	select {
	case c.initTaskChan <- initJob:
		return nil
	default:
		log.GetBaseLogger().Warnf("[RateLimit]AsyncRateLimitConnector initTaskChan full")
		return errors.New("buffer full")
	}
}

// 判断是否启动
func (c *AsyncRateLimitConnector) IsStart() error {
	if atomic.LoadInt32(&c.start) == 0 {
		c.lock.Lock()
		defer c.lock.Unlock()
		if atomic.LoadInt32(&c.start) == 0 {
			err := c.Start()
			if err != nil {
				log.GetBaseLogger().Errorf("[RateLimit]AsyncRateLimitConnector start error:%s", err.Error())
				return err
			}
		}
	}
	return nil
}

func (c *AsyncRateLimitConnector) GetStreamForUsing(hashKey string, op string) (*StreamWrapper, bool, error) {
	addr, _, err := c.getKeyHashAddr(hashKey)
	if err != nil {
		return nil, false, err
	}
	var streamMap *sync.Map
	var lastAddrMap *sync.Map
	if op == connector.OpKeyRateLimitAcquire {
		streamMap = &c.acquireStreamMap
		lastAddrMap = &c.acquireKeyAddrMap
	} else {
		log.GetBaseLogger().Warnf("unexpected op key:%s", op)
		return nil, false, errors.New(fmt.Sprintf("unexpected op key:%s", op))
	}

	changeAddr := false
	isNewAddr := false
	lastAddr, ok := lastAddrMap.Load(hashKey)
	if ok {
		if lastAddr != addr {
			changeAddr = true
			lastAddrMap.Store(hashKey, addr)
		}
	} else {
		isNewAddr = true
		lastAddrMap.Store(hashKey, addr)
	}
	var stream *StreamWrapper
	mV, exists := streamMap.Load(addr)
	if exists {
		stream = mV.(*StreamWrapper)
		if atomic.LoadInt32(&stream.canceled) == 1 {
			return nil, false, errors.New("stream canceled")
		}
	} else {
		stream, err = c.getStreamWrapper(hashKey, op)
		if err != nil {
			return nil, changeAddr, err
		}
		streamMap.Store(addr, stream)
		stream.keyAcquire()
		if op == connector.OpKeyRateLimitAcquire {
			go c.doReceiveAcquire(stream)
		}
		changeAddr = true
	}
	if isNewAddr {
		stream.keyAcquire()
	}
	if changeAddr {
		stream.keyAcquire()
		if v, ok := streamMap.Load(lastAddr); ok {
			sv := v.(*StreamWrapper)
			sv.keyRelease()
		}
	}
	return stream, changeAddr, nil
}

func (c *AsyncRateLimitConnector) doSyncRateLimitInit(initJob *WindowInitJob) error {
	request := initJob.request.(*rlimitpb.RateLimitRequest)
	addr, instance, err := c.getKeyHashAddr(request.GetKey().GetValue())
	if err != nil {
		return err
	}
	var stream *StreamWrapper
	sv, exist := c.rateLimitInitConnMap.Load(addr)
	if !exist {
		conn, err := c.connManager.ConnectByAddr(config.MetricCluster, addr, instance)
		if err != nil {
			return err
		}
		stream = &StreamWrapper{
			conn: conn,
		}
		c.rateLimitInitConnMap.Store(addr, stream)
	} else {
		stream = sv.(*StreamWrapper)
	}
	timeout := time.Millisecond * 500
	rLimitClient := rlimitpb.NewRateLimitGRPCClient(network.ToGRPCConn(stream.conn.Conn))
	reqID := connector.NextRateLimitInitReqID()
	ctx, cancel := connector.CreateHeaderContext(timeout, reqID)
	if cancel != nil {
		defer cancel()
	}
	if log.GetBaseLogger().IsLevelEnabled(log.DebugLog) {
		reqJson, _ := (&jsonpb.Marshaler{}).MarshalToString(request)
		log.GetBaseLogger().Debugf("[RateLimit]request to send is %s, opKey %s, connID %s", reqJson,
			request.GetKey().GetValue(), stream.conn.ConnID.String())
	}
	startTime := time.Now()
	atomic.StoreInt64(&stream.lastUseTime, startTime.Unix())
	resp, err := rLimitClient.InitializeQuota(ctx, request)
	if nil != err {
		return err
	}
	if model.IsSuccessResultCode(resp.GetCode().GetValue()) {
		log.GetBaseLogger().Debugf("[RateLimit]doSyncRateLimitInit done key:%s", request.GetKey().GetValue())
		duration := model.ToMilliSeconds(time.Since(startTime))
		initJob.win.OnResponse(nil, resp, duration/2, true)
		return nil
	} else {
		log.GetBaseLogger().Warnf("[RateLimit]doSyncRateLimitInit err resp:%s", resp.String())
		return errors.New(fmt.Sprintf("doSyncRateLimitInit err"))
	}
}

//异步上报使用的配额
func (c *AsyncRateLimitConnector) AsyncAcquire(req proto.Message, options interface{}) error {
	var err error
	err = c.IsStart()
	if err != nil {
		return err
	}
	request := req.(*rlimitpb.RateLimitRequest)
	if log.GetBaseLogger().IsLevelEnabled(log.DebugLog) {
		log.GetBaseLogger().Debugf("[RateLimit] AsyncRateLimitConnector Acquire request:%s", request.String())
	}
	sendReq := &SendRequest{
		reqType: connector.OpKeyRateLimitAcquire,
		request: request,
	}
	rWindow := options.(*quota.RateLimitWindow)

	key := request.GetKey().GetValue()
	stream, needReInit, err := c.GetStreamForUsing(key, connector.OpKeyRateLimitAcquire)
	if err != nil {
		log.GetBaseLogger().Debugf("[CheckConnection]traceCheck err:%s", err.Error())
		return err
	}
	if needReInit {
		log.GetBaseLogger().Debugf("[RateLimit]needReInit:%s", request.GetKey().GetValue())
		initJob := &WindowInitJob{
			win:     rWindow,
			reqType: connector.OpKeyRateLimitInit,
			request: rWindow.InitializeRequest(),
		}
		err = c.sendToInitChannel(initJob)
		if err != nil {
			log.GetBaseLogger().Warnf("[RateLimit]isNewAddr should reInit %s but error:%s",
				request.GetKey().GetValue(), err.Error())
		}
	}
	sendReq.stream = stream
	sendReq.win = rWindow
	err = c.sendToLocalChannel(sendReq)
	if err != nil {
		return err
	} else {
		stream.acquire()
		key := request.GetKey().GetValue()
		timeNow := clock.GetClock().Now().Unix()
		if v, ok := c.acquireWindowMap.Load(key); ok {
			rw := v.(*WindowRecord)
			if rw.win.GetWindowInitTime() != rWindow.GetWindowInitTime() {
				record := &WindowRecord{
					win:      rWindow,
					lastTime: timeNow,
				}
				c.acquireWindowMap.Delete(key)
				c.acquireWindowMap.Store(key, record)
			} else {
				rw.lastTime = timeNow
			}
		} else {
			record := &WindowRecord{
				win:      rWindow,
				lastTime: timeNow,
			}
			c.acquireWindowMap.Store(key, record)
		}

	}
	return nil
}

// 清理过期的window
func (c *AsyncRateLimitConnector) ClearExpireWindow(key string) error {
	c.acquireWindowMap.Delete(key)
	lastAddrMap := &c.acquireKeyAddrMap
	lastAddr, ok := lastAddrMap.Load(key)
	if !ok {
		return nil
	}
	lastAddrMap.Delete(key)
	mV, ok := c.acquireStreamMap.Load(lastAddr)
	if ok {
		_ = mV
		stream := mV.(*StreamWrapper)
		stream.keyRelease()
	}
	return nil
}
