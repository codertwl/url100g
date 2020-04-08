package logic

import (
	"bufio"
	"container/heap"
	"fmt"
	"io"
	"os"
	"sort"
	"strconv"
	"strings"
)

type FILE struct {
	fp   *os.File
	w    *bufio.Writer
	r    *bufio.Reader
	open bool
}

type UC struct {
	url string
	cnt int64
}

func (u *UC) String() string {
	return fmt.Sprintf("url:%s cnt:%d", u.url, u.cnt)
}

type HEAP []*UC

func (m HEAP) Len() int            { return len(m) }
func (m HEAP) Less(i, j int) bool  { return m[i].cnt < m[j].cnt }
func (m HEAP) Swap(i, j int)       { m[i], m[j] = m[j], m[i] }
func (m *HEAP) Push(x interface{}) { *m = append(*m, x.(*UC)) }
func (m *HEAP) Pop() interface{} {
	if len(*m) == 0 {
		return &UC{}
	}
	h := (*m)[len(*m)-1]
	*m = (*m)[:len(*m)-1]
	return h
}

var mapSep map[int]*FILE = make(map[int]*FILE)

func GetSepFileDir(outPath string) string {
	return fmt.Sprintf("%s/seps", outPath)
}

func GetSortFileDir(outPath string) string {
	return fmt.Sprintf("%s/sorts", outPath)
}

func getSepFileFullPath(path string, saveNo int) string {
	return fmt.Sprintf("%s/sep_%d", path, saveNo)
}

func getSortFileFullPath(path string, saveNo int) string {
	return fmt.Sprintf("%s/sort_%d", path, saveNo)
}

func getSaveNo(url string, sepMax int) int {

	hash := uint64(2166136261)

	for i := 0; i < len(url); i++ {
		hash *= 16777619
		hash ^= uint64(url[i])
	}

	return int(hash % uint64(sepMax))
}

func appendLineToFile(fp *FILE, line string) error {

	_, err := fp.w.WriteString(line)
	if err != nil {
		panic("appendLineToFile err:" + err.Error())
		return err
	}

	return nil
}

func getSepFileFpBySaveNo(saveNo int, sepPath string) (*FILE, error) {

	f, ok := mapSep[saveNo]
	if ok {
		return f, nil
	}

	//create
	fp, err := os.Create(getSepFileFullPath(sepPath, saveNo))
	if err != nil {
		fmt.Printf("create sep file error: %v\n", err)
		return nil, err
	}

	f = &FILE{fp: fp, w: bufio.NewWriter(fp), open: true}
	mapSep[saveNo] = f

	return f, nil
}

func SepBigFile(bigFile, outPath string, sepMax, topN int) {

	bFp, err := os.OpenFile(bigFile, os.O_RDONLY, 0666)
	if err != nil {
		fmt.Println("open bigFile err:", err)
		return
	}
	defer bFp.Close()

	buf := bufio.NewReader(bFp)
	for {
		line, err := buf.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				fmt.Println("read bigFile over...")
				break
			} else {
				panic("read bigFile line err:" + err.Error())
			}
		}

		saveNo := getSaveNo(line, sepMax)
		sFp, err := getSepFileFpBySaveNo(saveNo, GetSepFileDir(outPath))
		if err != nil {
			panic("getSepFileFpBySaveNo err:" + err.Error())
		}

		err = appendLineToFile(sFp, line)
		if err != nil {
			panic("appendLineToFile line:" + err.Error())
		}
	}

	for _, v := range mapSep {
		v.w.Flush()
		v.fp.Close()
		v.open = false
	}

	sortSepFile(outPath, topN)
}

func saveSortFile(mUrlCount map[string]int64, saveNo, topN int, sortPath string) error {

	var listUC []*UC
	for k, v := range mUrlCount {
		listUC = append(listUC, &UC{url: k, cnt: v})
	}

	sort.Slice(listUC, func(i, j int) bool {
		return listUC[i].cnt > listUC[j].cnt
	})

	fp, err := os.Create(getSortFileFullPath(sortPath, saveNo))
	if err != nil {
		panic("create sort file error:" + err.Error())
		return err
	}
	defer fp.Close()

	w := bufio.NewWriter(fp)
	for k, v := range listUC {

		if k >= topN {
			break
		}

		line := fmt.Sprintf("%s %d\n", strings.Trim(v.url, "\n"), v.cnt)
		n, err := w.WriteString(line)
		if err != nil {
			panic(fmt.Sprintf("write:%d sortfile err:%s", n, err))
		}
	}

	w.Flush()

	return nil
}

func sortSepFile(outPath string, topN int) {

	for k, _ := range mapSep {

		fp, err := os.Open(getSepFileFullPath(GetSepFileDir(outPath), k))
		if err != nil {
			panic("open sep file error:" + err.Error())
		}

		r := bufio.NewReader(fp)
		mUrlCount := make(map[string]int64)

		for {
			line, err := r.ReadString('\n')
			if err != nil {
				if err == io.EOF {
					break
				} else {
					panic(fmt.Sprintf("read sepfile:%d line err:%s", k, err))
				}
			}
			mUrlCount[line]++
		}

		fp.Close()
		saveSortFile(mUrlCount, k, topN, GetSortFileDir(outPath))
	}
}

func TopN(outPath string, n int) {

	heapN := make(HEAP, 0, n)
	heap.Init(&heapN)

	total := 0
	for len(mapSep) > 0 {
		for k, v := range mapSep {

			if !v.open {
				fp, err := os.Open(getSortFileFullPath(GetSortFileDir(outPath), k))
				if err != nil {
					panic("open sep file error:" + err.Error())
				}
				v = &FILE{fp: fp, r: bufio.NewReader(fp), open: true}
				mapSep[k] = v
			}

			line, err := v.r.ReadString('\n')
			if err != nil {
				if err == io.EOF {
					v.fp.Close()
					delete(mapSep, k)
					continue
				} else {
					panic(fmt.Sprintf("read sepfile:%d line err:%s", k, err))
				}
			}

			seps := strings.Split(line, " ")
			if len(seps) != 2 {
				panic("sep line wrong:" + line)
			}
			cnt, _ := strconv.ParseInt(strings.Trim(seps[1], "\n"), 10, 64)

			heap.Push(&heapN, &UC{url: seps[0], cnt: cnt})
			if heapN.Len() > n {
				heap.Pop(&heapN)
			}
			total++
		}
	}

	ret := make([]*UC, 0, heapN.Len())
	for heapN.Len() > 0 {
		t := heap.Pop(&heapN)
		ucT := t.(*UC)
		ret = append(ret, ucT)
	}

	for i := len(ret) - 1; i >= 0; i-- {
		fmt.Println(len(ret)-i, " : ", ret[i].url, " --------- ", ret[i].cnt)
	}

}
