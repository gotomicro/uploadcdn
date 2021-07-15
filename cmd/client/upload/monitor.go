package upload

import (
	"fmt"
	"strings"
	"sync/atomic"
	"time"
)

const (
	normalExit = iota
	errExit
)

var processTickInterval int64 = 5

var clearStrLen int = 0
var clearStr string = strings.Repeat(" ", clearStrLen)

func getClearStr(str string) string {
	if clearStrLen <= len(str) {
		clearStrLen = len(str)
		return fmt.Sprintf("\r%s", str)
	}
	clearStr = strings.Repeat(" ", clearStrLen)
	return fmt.Sprintf("\r%s\r%s", clearStr, str)
}

type CPMonitorSnap struct {
	transferSize  int64
	skipSize      int64
	dealSize      int64
	fileNum       int64
	dirNum        int64
	skipNum       int64
	skipNumDir    int64
	errNum        int64
	okNum         int64
	dealNum       int64
	duration      int64
	incrementSize int64
}

type Monitor struct {
	totalSize      int64
	totalNum       int64
	transferSize   int64
	skipSize       int64
	dealSize       int64
	fileNum        int64
	dirNum         int64
	skipNum        int64
	skipNumDir     int64
	errNum         int64
	lastSnapSize   int64
	tickDuration   int64
	seekAheadError error
	seekAheadEnd   bool
	finish         bool
	_              uint32 //Add padding to make sure the next data 64bits alignment
	lastSnapTime   time.Time
}

func (m *Monitor) init() {
	m.totalSize = 0
	m.totalNum = 0
	m.seekAheadEnd = false
	m.seekAheadError = nil
	m.transferSize = 0
	m.skipSize = 0
	m.dealSize = 0
	m.fileNum = 0
	m.dirNum = 0
	m.skipNum = 0
	m.errNum = 0
	m.finish = false
	m.lastSnapSize = 0
	m.lastSnapTime = time.Now()
	m.tickDuration = processTickInterval * int64(time.Second)
}

func (m *Monitor) setScanError(err error) {
	m.seekAheadError = err
	m.seekAheadEnd = true
}

func (m *Monitor) updateScanNum(num int64) {
	m.totalNum = m.totalNum + num
}

func (m *Monitor) updateScanSizeNum(size, num int64) {
	m.totalSize = m.totalSize + size
	m.totalNum = m.totalNum + num
}

func (m *Monitor) setScanEnd() {
	m.seekAheadEnd = true
}

func (m *Monitor) updateTransferSize(size int64) {
	atomic.AddInt64(&m.transferSize, size)
}

func (m *Monitor) updateDealSize(size int64) {
	atomic.AddInt64(&m.dealSize, size)
}

func (m *Monitor) updateFile(size, num int64) {
	atomic.AddInt64(&m.fileNum, num)
	atomic.AddInt64(&m.transferSize, size)
	atomic.AddInt64(&m.dealSize, size)
}

func (m *Monitor) updateDir(size, num int64) {
	atomic.AddInt64(&m.dirNum, num)
	atomic.AddInt64(&m.transferSize, size)
	atomic.AddInt64(&m.dealSize, size)
}

func (m *Monitor) updateSkip(size, num int64) {
	atomic.AddInt64(&m.skipNum, num)
	atomic.AddInt64(&m.skipSize, size)
}

func (m *Monitor) updateSkipDir(num int64) {
	atomic.AddInt64(&m.skipNumDir, num)
}

func (m *Monitor) updateErr(size, num int64) {
	atomic.AddInt64(&m.errNum, num)
	atomic.AddInt64(&m.transferSize, size)
}

func (m *Monitor) getSnapshot() *CPMonitorSnap {
	var snap CPMonitorSnap
	snap.transferSize = m.transferSize
	snap.skipSize = m.skipSize
	snap.dealSize = m.dealSize + snap.skipSize
	snap.fileNum = m.fileNum
	snap.dirNum = m.dirNum
	snap.skipNum = m.skipNum
	snap.errNum = m.errNum
	snap.okNum = snap.fileNum + snap.dirNum + snap.skipNum
	snap.dealNum = snap.okNum + snap.errNum
	snap.skipNumDir = m.skipNumDir
	now := time.Now()
	snap.duration = now.Sub(m.lastSnapTime).Nanoseconds()

	return &snap
}

func (m *Monitor) progressBar(finish bool, exitStat int) string {
	if m.finish {
		return ""
	}
	m.finish = m.finish || finish
	if !finish {
		return m.getProgressBar()
	}
	return m.getFinishBar(exitStat)
}

func (m *Monitor) getProgressBar() string {
	mu.RLock()
	defer mu.RUnlock()

	snap := m.getSnapshot()
	if snap.duration < m.tickDuration {
		return ""
	} else {
		m.lastSnapTime = time.Now()
		snap.incrementSize = m.transferSize - m.lastSnapSize
		m.lastSnapSize = snap.transferSize
	}

	if m.seekAheadEnd && m.seekAheadError == nil {
		return getClearStr(fmt.Sprintf("Total num: %d, size: %s. Dealed num: %d%s%s, Progress: %.3f%s, Speed: %.2fKB/s", m.totalNum, getSizeString(m.totalSize), snap.dealNum, m.getDealNumDetail(snap), m.getDealSizeDetail(snap), m.getPrecent(snap), "%%", m.getSpeed(snap)))
	}
	scanNum := max(m.totalNum, snap.dealNum)
	scanSize := max(m.totalSize, snap.dealSize)
	return getClearStr(fmt.Sprintf("Scanned num: %d, size: %s. Dealed num: %d%s%s, Speed: %.2fKB/s.", scanNum, getSizeString(scanSize), snap.dealNum, m.getDealNumDetail(snap), m.getDealSizeDetail(snap), m.getSpeed(snap)))
}

func (m *Monitor) getFinishBar(exitStat int) string {
	if exitStat == normalExit {
		return m.getWholeFinishBar()
	}
	return m.getDefeatBar()
}

func (m *Monitor) getWholeFinishBar() string {
	snap := m.getSnapshot()
	if m.seekAheadEnd && m.seekAheadError == nil {
		if snap.errNum == 0 {
			return getClearStr(fmt.Sprintf("Succeed: Total num: %d, size: %s. OK num: %d%s%s.\n", m.totalNum, getSizeString(m.totalSize), snap.okNum, m.getDealNumDetail(snap), m.getSkipSize(snap)))
		}
		return getClearStr(fmt.Sprintf("FinishWithError: Total num: %d, size: %s. Error num: %d. OK num: %d%s%s.\n", m.totalNum, getSizeString(m.totalSize), snap.errNum, snap.okNum, m.getOKNumDetail(snap), m.getSizeDetail(snap)))
	}
	scanNum := max(m.totalNum, snap.dealNum)
	if snap.errNum == 0 {
		return getClearStr(fmt.Sprintf("Succeed: Total num: %d, size: %s. OK num: %d%s%s.\n", scanNum, getSizeString(snap.dealSize), snap.okNum, m.getDealNumDetail(snap), m.getSkipSize(snap)))
	}
	return getClearStr(fmt.Sprintf("FinishWithError: Scanned %d %s. Error num: %d. OK num: %d%s%s.\n", scanNum, m.getSubject(), snap.errNum, snap.okNum, m.getOKNumDetail(snap), m.getSizeDetail(snap)))
}

func (m *Monitor) getDefeatBar() string {
	snap := m.getSnapshot()
	if m.seekAheadEnd && m.seekAheadError == nil {
		return getClearStr(fmt.Sprintf("Total num: %d, size: %s. Dealed num: %d%s%s. When error happens.\n", m.totalNum, getSizeString(m.totalSize), snap.okNum, m.getOKNumDetail(snap), m.getSizeDetail(snap)))
	}
	scanNum := max(m.totalNum, snap.dealNum)
	return getClearStr(fmt.Sprintf("Scanned %d %s. Dealed num: %d%s%s. When error happens.\n", scanNum, m.getSubject(), snap.okNum, m.getOKNumDetail(snap), m.getSizeDetail(snap)))
}

func (m *Monitor) getSubject() string {
	return "files"
}

func (m *Monitor) getDealNumDetail(snap *CPMonitorSnap) string {
	return m.getNumDetail(snap, true)
}

func (m *Monitor) getOKNumDetail(snap *CPMonitorSnap) string {
	return m.getNumDetail(snap, false)
}

func (m *Monitor) getNumDetail(snap *CPMonitorSnap, hasErr bool) string {
	if !hasErr && snap.okNum == 0 {
		return ""
	}
	strList := []string{}
	if hasErr && snap.errNum != 0 {
		strList = append(strList, fmt.Sprintf("Error %d %s", snap.errNum, m.getSubject()))
	}
	if snap.fileNum != 0 {
		strList = append(strList, fmt.Sprintf("%s %d %s", m.getOPStr(), snap.fileNum, m.getSubject()))
	}
	if snap.dirNum != 0 {
		str := fmt.Sprintf("%d directories", snap.dirNum)
		if snap.fileNum == 0 {
			str = fmt.Sprintf("%s %d directories", m.getOPStr(), snap.dirNum)
		}
		strList = append(strList, str)
	}
	if snap.skipNum != 0 {
		strList = append(strList, fmt.Sprintf("skip %d %s", snap.skipNum, m.getSubject()))
	}
	if snap.skipNumDir != 0 {
		strList = append(strList, fmt.Sprintf("skip %d directory", snap.skipNumDir))
	}

	if len(strList) == 0 {
		return ""
	}
	return fmt.Sprintf("(%s)", strings.Join(strList, ", "))
}

func (m *Monitor) getSpeed(snap *CPMonitorSnap) float64 {
	return (float64(snap.incrementSize) / 1024) / (float64(snap.duration) * 1e-9)
}

func (m *Monitor) getOPStr() string {
	return "upload"
}

func (m *Monitor) getDealSizeDetail(snap *CPMonitorSnap) string {
	return fmt.Sprintf(", OK size: %s", getSizeString(snap.dealSize))
}

func (m *Monitor) getSkipSize(snap *CPMonitorSnap) string {
	if snap.skipSize != 0 {
		return fmt.Sprintf(", Skip size: %s", getSizeString(snap.skipSize))
	}
	return ""
}

func (m *Monitor) getSizeDetail(snap *CPMonitorSnap) string {
	if snap.skipSize == 0 {
		return fmt.Sprintf(", Transfer size: %s", getSizeString(snap.transferSize))
	}
	if snap.transferSize == 0 {
		return fmt.Sprintf(", Skip size: %s", getSizeString(snap.skipSize))
	}
	return fmt.Sprintf(", OK size: %s(transfer: %s, skip: %s)", getSizeString(snap.transferSize+snap.skipSize), getSizeString(snap.transferSize), getSizeString(snap.skipSize))
}

func (m *Monitor) getPrecent(snap *CPMonitorSnap) float64 {
	if m.seekAheadEnd && m.seekAheadError == nil {
		if m.totalSize != 0 {
			return float64((snap.dealSize)*100.0) / float64(m.totalSize)
		}
		if m.totalNum != 0 {
			return float64((snap.dealNum)*100.0) / float64(m.totalNum)
		}
		return 100
	}
	return 0
}
