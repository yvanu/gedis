package aof

import (
	gedisconn "gedis/gedis/conn"
	"gedis/gedis/parser"
	"gedis/gedis/proto"
	"gedis/iface"
	"gedis/tool/logger"
	"io"
	"os"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

const (
	aofChanSize = 1 << 20
)

const (
	FsyncAlways   = "always"
	FsyncEverySec = "everysec"
	FsyncNo       = "no"
)

type Command = [][]byte

type aofRecord struct {
	dbIndex int
	command Command
}

// AOF  Append Only File
type AOF struct {
	aofFile     *os.File
	aofFileName string
	// 刷盘间隔
	aofFsync    string
	lastDBIndex int

	aofChan chan aofRecord
	mu      sync.Mutex

	aofFinished chan struct{}
	close       chan struct{}
	atomicClose atomic.Bool

	engine iface.Engine
}

func NewAOF(aofFileName string, engine iface.Engine, fsync string) (*AOF, error) {
	aof := &AOF{}
	aof.aofFileName = aofFileName
	aof.aofFsync = strings.ToLower(fsync)
	aof.aofChan = make(chan aofRecord)
	aof.close = make(chan struct{})
	aof.aofFinished = make(chan struct{})
	aof.engine = engine
	aof.atomicClose.Store(false)

	// 打开aof文件
	var err error
	aof.aofFile, err = os.OpenFile(aofFileName, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return nil, err
	}

	if aof.aofFsync == FsyncEverySec {
		aof.fsyncEverySec()
	}
	go aof.watchAofChan()
	return aof, nil
}

func (a *AOF) watchAofChan() {
	for record := range a.aofChan {
		a.writeAofRecord(record)
	}
	a.aofFinished <- struct{}{}
}

func (a *AOF) writeAofRecord(record aofRecord) {
	a.mu.Lock()
	defer a.mu.Unlock()

	if record.dbIndex != a.lastDBIndex {
		// 记录该次命令的db
		selectCmd := Command{[]byte("select"), []byte(strconv.Itoa(record.dbIndex))}
		data := proto.NewMultiBulkReply(selectCmd).Bytes()
		_, err := a.aofFile.Write(data)
		if err != nil {
			logger.Warn("write aof record error:", err)
			return
		}
		a.lastDBIndex = record.dbIndex
	}
	// 记录实际的命令
	data := proto.NewMultiBulkReply(record.command).Bytes()
	_, err := a.aofFile.Write(data)
	if err != nil {
		logger.Warn("write aof record error:", err)
	}
	logger.Debugf("write aof command: %q", data)
	if a.aofFsync == FsyncAlways {
		a.aofFile.Sync()
	}
}

func (a *AOF) fsyncEverySec() {
	ticker := time.NewTicker(time.Second)
	go func() {
		for {
			select {
			case <-ticker.C:
				a.Fsync()
			case <-a.close:
				return
			}
		}
	}()
}

func (a *AOF) Fsync() {
	a.mu.Lock()
	defer a.mu.Unlock()
	err := a.aofFile.Sync()
	if err != nil {
		logger.Warn("fsync aof error:", err)
	}
}

func (a *AOF) LoadAof(maxBytes int) {
	// 展示关闭aof记录
	a.atomicClose.Store(true)
	defer func() {
		a.atomicClose.Store(false)
	}()
	file, err := os.Open(a.aofFileName)
	if err != nil {
		logger.Error("open aof file error:", err)
		return
	}
	defer file.Close()
	file.Seek(0, io.SeekStart)

	var reader io.Reader
	if maxBytes > 0 {
		reader = io.LimitReader(file, int64(maxBytes))
	} else {
		reader = file
	}

	ch := parser.ParseStream(reader)
	virtualConn := gedisconn.NewVirtualConnection()

	for payload := range ch {
		if payload.Err != nil {
			if payload.Err == io.EOF {
				break
			}
			logger.Error("parse aof record error:", payload.Err)
			continue
		}
		if payload.Reply == nil {
			logger.Warn("empty payload")
			continue
		}
		reply, ok := payload.Reply.(*proto.MultiBulkReply)
		if !ok {
			logger.Error("is not MultiBulkReply")
			continue
		}
		ret := a.engine.Exec(virtualConn, reply.Command)
		if proto.IsErrReply(ret) {
			logger.Error("exec aof command error:", ret)
		}
		if strings.ToLower(string(reply.Command[0])) == "select" {
			dbIndex, err := strconv.Atoi(string(reply.Command[1]))
			if err == nil {
				a.lastDBIndex = dbIndex
			}
		}
	}
}

func (a *AOF) SaveGedisCommand(index int, command [][]byte) {
	if a.atomicClose.Load() {
		return
	}
	if a.aofFsync == FsyncAlways {
		a.writeAofRecord(aofRecord{
			dbIndex: index,
			command: command,
		})
		return
	}
	a.aofChan <- aofRecord{
		dbIndex: index,
		command: command,
	}
}
