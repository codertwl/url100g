package main

import (
	"flag"
	"fmt"
	"github.com/codertwl/url100g/logic"
	"os"
	"time"
)

const ()

func main() {

	sepMax, bigFile, outPath, n := 11731, "", "", 100

	//parse param
	flag.IntVar(&sepMax, "s", 11731, "")
	flag.StringVar(&bigFile, "b", "", "")
	flag.StringVar(&outPath, "o", "./tmp", "default ./tmp")
	flag.IntVar(&n, "n", 100, "")
	flag.Parse()

	if sepMax <= 0 || bigFile == "" || outPath == "" || n <= 0 {
		fmt.Println("invalid params")
		return
	}

	err := os.MkdirAll(logic.GetSepFileDir(outPath), os.ModePerm)
	if err != nil {
		fmt.Println("mkdir seps err:", err)
		return
	}

	err = os.MkdirAll(logic.GetSortFileDir(outPath), os.ModePerm)
	if err != nil {
		fmt.Println("mkdir sorts err:", err)
		return
	}

	if _, err = os.Stat(bigFile); err != nil /*os.IsNotExist(err)*/ {
		fmt.Println("check file err:", err)
		return
	}

	//seperate big file
	start := time.Now().Unix()
	logic.SepBigFile(bigFile, outPath, sepMax, n)
	fmt.Println("sep over...cost ", time.Now().Unix()-start, " sec")

	//get top n
	start = time.Now().Unix()
	logic.TopN(outPath, n)
	fmt.Println("topn over...cost ", time.Now().Unix()-start, " sec")
}
