package server

import (
	"bufio"
	"encoding/json"
	"log"
	"net"
	"sync"
	"time"
)

type StreamCloser func(s *Stream)

type Stream struct {
	id           string
	locker       sync.Mutex
	conn         net.Conn
	closer       sync.Once
	reader       *bufio.Reader
	isActive     chan bool
	latestActive int64
	cbCloser     StreamCloser
	waiter       map[uint64]chan<- Packer
}

func NewStream(conn net.Conn, cb StreamCloser) *Stream {
	s := &Stream{
		conn:         conn,
		reader:       bufio.NewReader(conn),
		isActive:     make(chan bool),
		cbCloser:     cb,
		latestActive: time.Now().Unix(),
		waiter:       make(map[uint64]chan<- Packer),
	}

	s.init()

	return s
}

func (self *Stream) init() {
	log.Println(self.conn.RemoteAddr(), "is close")

	// 启动活动检测
	go self.checkIsActive()
	// 启动接收流程
}

// 定时检测是否处理活动状态
func (self *Stream) checkIsActive() {
	defer self.Close()

	tick := time.NewTicker(time.Second * 5)
	defer tick.Stop()

	for {
		select {
		case _, _ = <-self.isActive:
			break
		case _, _ = <-tick.C:
		}

		// 判断是否过期
		now := time.Now().Unix()
		active := self.latestActive + 10
		if active < now {
			log.Println(self.conn.RemoteAddr(), "time out")
			return
		}
	}
}

func (self *Stream) GetID() string {
	return self.id
}

func (self *Stream) SetID(id string) {
	self.id = id
}

func (self *Stream) RemoteAddr() net.Addr {
	return self.conn.RemoteAddr()
}

// 关闭函数
func (self *Stream) Close() error {
	self.closer.Do(func() {
		close(self.isActive)
		// 关闭网络
		self.conn.Close()
		// 已经关闭
		log.Println(self.conn.RemoteAddr(), "is close")
		// 调用回调
		self.cbCloser(self)
	})

	return nil
}

// 写入函数
func (self *Stream) Write(buf []byte) (int, error) {
	self.locker.Lock()
	defer self.locker.Unlock()

	n, err := self.conn.Write(buf)
	if nil != err {
		return 0, err
	}

	self.latestActive = time.Now().Unix()
	return n, nil
}

// 关闭函数
func (self *Stream) Read(buf []byte) (int, error) {
	n, err := self.reader.Read(buf)
	if nil != err {
		return 0, err
	}

	self.latestActive = time.Now().Unix()
	return n, nil
}

// 写入一帧数据
func (self *Stream) WritePack(pkt interface{}) error {
	buf, err := json.Marshal(pkt)
	if nil != err {
		return err
	}

	for len(buf) > 0 {
		n, err := self.Write(buf)
		if nil != err {
			return err
		}

		buf = buf[n:]
	}

	return nil
}

// 读取一帧数据
func (self *Stream) ReadPack(pkt interface{}) error {
	line, _, err := self.reader.ReadLine()
	if nil != err {
		return err
	}

	self.latestActive = time.Now().Unix()
	return json.Unmarshal(line, pkt)
}
