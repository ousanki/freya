package utils

import (
	"errors"
	"freya/global"
	"sync"
	"time"
)

const (
	numberBitsLow  uint8 = 10 // 表示每个集群下的每个节点，1毫秒内可生成的id序号的二进制位 对应上图中的最后一段
	workerBitsLow  uint8 = 8 // 每台机器(节点)的ID位数 10位最大可以有2^8=256个节点数 即每毫秒可生成 2^10-1=1023个唯一ID 对应上图中的倒数第二段
	workerMaxLow   int64 = -1 ^ (-1 << workerBitsLow) // 节点ID的最大值，用于防止溢出
	numberMaxLow   int64 = -1 ^ (-1 << numberBitsLow) // 同上，用来表示生成id序号的最大值
	timeShiftLow   uint8 = workerBitsLow + numberBitsLow // 时间戳向左的偏移量
	workerShiftLow uint8 = numberBitsLow // 节点ID向左的偏移量
	epochLow int64 = 1525705533 // 这个是我在写epoch这个常量时的时间戳(秒)
)

// 定义一个woker工作节点所需要的基本参数
type WorkerLow struct {
	mu        sync.Mutex // 添加互斥锁 确保并发安全
	timestamp int64      // 记录上一次生成id的时间戳
	workerId  int64      // 该节点的ID
	number    int64      // 当前毫秒已经生成的id序列号(从0开始累加) 1秒内最多生成4096个ID
}

var workeLow *WorkerLow

func InitIdWorkerLow() {
	var e error
	workeLow, e = newWorkerLow(int64(global.G.ServerId))
	if e != nil {
		panic(e.Error())
	}
}

// 实例化一个工作节点
// workerId 为当前节点的id
func newWorkerLow(workerId int64) (*WorkerLow, error) {
	// 要先检测workerId是否在上面定义的范围内
	if workerId < 0 || workerId > workerMaxLow {
		return nil, errors.New("WorkerLow ID excess of quantity")
	}
	// 生成一个新节点
	return &WorkerLow{
		timestamp: 0,
		workerId:  workerId,
		number:    0,
	}, nil
}

// 生成方法一定要挂载在某个woker下，这样逻辑会比较清晰 指定某个节点生成id
func (w *WorkerLow) nextIDLow() int64 {
	// 获取id最关键的一点 加锁 加锁 加锁
	w.mu.Lock()
	defer w.mu.Unlock() // 生成完成后记得 解锁 解锁 解锁

	// 获取生成时的时间戳
	now := time.Now().Unix() // 获取当前秒
	if w.timestamp == now {
		w.number++
		// 这里要判断，当前工作节点是否在1秒内已经生成numberMax个ID
		if w.number > numberMaxLow {
			// 如果当前工作节点在1秒内生成的ID已经超过上限 需要等待1秒再继续生成
			for now <= w.timestamp {
				now = time.Now().Unix()
			}
		}
	} else {
		// 如果当前时间与工作节点上一次生成ID的时间不一致 则需要重置工作节点生成ID的序号
		w.number = 0
		// 下面这段代码看到很多前辈都写在if外面，无论节点上次生成id的时间戳与当前时间是否相同 都重新赋值  这样会增加一丢丢的额外开销 所以我这里是选择放在else里面
		w.timestamp = now // 将机器上一次生成ID的时间更新为当前时间
	}

	ID := int64((now - epochLow) << timeShiftLow | (w.workerId << workerShiftLow) | (w.number))

	return ID
}

func NextIDLow() uint64 {
	return uint64(workeLow.nextIDLow())
}

func WorkerIDLow(nId int64) int64 {
	ID := nId << (64 - timeShiftLow) >> (64 - workerBitsLow)
	return ID
}
